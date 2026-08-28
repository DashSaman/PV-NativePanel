package runtimeagent

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DashSaman/PV-NaivePanel/internal/naiveruntime"
	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

const (
	productionCaddyfilePath = "/etc/caddy/Caddyfile"
	productionCaddyBinary   = "/usr/local/bin/caddy"
	productionServiceName   = "caddy-naive.service"
	productionBackupRoot    = "/var/backups/pvnaive/caddy"
	commandTimeout          = 20 * time.Second
)

type commandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("fixed command failed: %w", err)
	}
	return output, nil
}

type operatorConfig struct {
	caddyfilePath string
	caddyBinary   string
	serviceName   string
	backupRoot    string
}

type FixedOperator struct {
	config operatorConfig
	runner commandRunner
	now    func() time.Time
	random io.Reader
}

type fileMetadata struct {
	mode os.FileMode
	uid  int
	gid  int
}

type serviceSnapshot struct {
	mainPID   int
	nRestarts int
}

func NewOperator() (*FixedOperator, error) {
	return newOperator(operatorConfig{
		caddyfilePath: productionCaddyfilePath,
		caddyBinary:   productionCaddyBinary,
		serviceName:   productionServiceName,
		backupRoot:    productionBackupRoot,
	}, execCommandRunner{}, time.Now, rand.Reader)
}

func newOperator(config operatorConfig, runner commandRunner, now func() time.Time, random io.Reader) (*FixedOperator, error) {
	if config.caddyfilePath == "" || config.caddyBinary == "" || config.serviceName == "" || config.backupRoot == "" {
		return nil, errors.New("runtimeagent: incomplete fixed operator config")
	}
	if !filepath.IsAbs(config.caddyfilePath) || !filepath.IsAbs(config.caddyBinary) || !filepath.IsAbs(config.backupRoot) {
		return nil, errors.New("runtimeagent: operator paths must be absolute")
	}
	if runner == nil || now == nil || random == nil {
		return nil, errors.New("runtimeagent: operator dependencies are required")
	}
	return &FixedOperator{config: config, runner: runner, now: now, random: random}, nil
}

func (o *FixedOperator) Health(ctx context.Context) (HealthResponse, error) {
	current, _, err := o.readCurrent()
	if err != nil {
		return HealthResponse{Status: "degraded"}, nil
	}
	if _, err := naiveruntime.InspectCaddyfile(current); err != nil {
		return HealthResponse{Status: "degraded"}, nil
	}
	if _, err := o.snapshotService(ctx); err != nil {
		return HealthResponse{Status: "degraded"}, nil
	}
	return HealthResponse{Status: "ok"}, nil
}

func (o *FixedOperator) Inspect(_ context.Context) (InspectResponse, error) {
	current, _, err := o.readCurrent()
	if err != nil {
		return InspectResponse{}, err
	}
	inspection, err := naiveruntime.InspectCaddyfile(current)
	if err != nil {
		return InspectResponse{}, fmt.Errorf("runtimeagent: inspect Caddyfile: %w", err)
	}
	credentials, err := inspection.CredentialsForImport()
	if err != nil {
		return InspectResponse{}, fmt.Errorf("runtimeagent: build import credential view: %w", err)
	}
	response := InspectResponse{
		CaddySHA256: sha256Hex(current),
		Credentials: make([]InspectCredential, 0, len(credentials)),
	}
	for _, credential := range credentials {
		response.Credentials = append(response.Credentials, InspectCredential{
			Username: credential.Username,
			Password: credential.Password(),
		})
	}
	return response, nil
}

func (o *FixedOperator) Validate(ctx context.Context, request ValidateRequest) (ValidateResponse, error) {
	candidate, metadata, err := o.prepareCandidate(request.ExpectedCaddySHA256, request.Desired)
	if err != nil {
		return ValidateResponse{}, err
	}
	candidatePath, err := o.writeTemp(candidate, metadata)
	if err != nil {
		return ValidateResponse{}, err
	}
	defer os.Remove(candidatePath)
	if err := o.validatePath(ctx, candidatePath); err != nil {
		return ValidateResponse{}, err
	}
	return ValidateResponse{CandidateSHA256: sha256Hex(candidate)}, nil
}

func (o *FixedOperator) Apply(ctx context.Context, request ApplyRequest) (ApplyResponse, error) {
	current, metadata, err := o.readCurrent()
	if err != nil {
		return ApplyResponse{}, err
	}
	currentSHA := sha256Hex(current)
	if request.ExpectedCaddySHA256 != currentSHA {
		return ApplyResponse{}, errors.New("runtimeagent: current Caddyfile SHA does not match expected revision")
	}
	if !validMutationInput(request.ExpectedCaddySHA256, request.Desired) {
		return ApplyResponse{}, errors.New("runtimeagent: invalid desired state")
	}
	candidate, err := renderDesired(current, request.Desired)
	if err != nil {
		return ApplyResponse{}, err
	}
	candidatePath, err := o.writeTemp(candidate, metadata)
	if err != nil {
		return ApplyResponse{}, err
	}
	defer os.Remove(candidatePath)
	if err := o.validatePath(ctx, candidatePath); err != nil {
		return ApplyResponse{}, err
	}

	before, err := o.snapshotService(ctx)
	if err != nil {
		return ApplyResponse{}, err
	}
	backupID, err := o.createBackup(current)
	if err != nil {
		return ApplyResponse{}, err
	}
	if err := o.installTemp(candidatePath); err != nil {
		return ApplyResponse{}, err
	}

	after, err := o.reloadAndVerify(ctx, before)
	if err == nil {
		installed, _, readErr := o.readCurrent()
		if readErr != nil {
			err = readErr
		} else if sha256Hex(installed) != sha256Hex(candidate) {
			err = errors.New("runtimeagent: installed Caddyfile SHA differs from validated candidate")
		}
	}
	if err != nil {
		compensationErr := o.restoreAndReload(ctx, current, metadata, before)
		if compensationErr != nil {
			return ApplyResponse{}, fmt.Errorf("runtimeagent: apply failed and compensation failed: %v; compensation: %w", err, compensationErr)
		}
		return ApplyResponse{}, fmt.Errorf("runtimeagent: apply failed and was rolled back: %w", err)
	}

	return ApplyResponse{
		PreviousSHA256: currentSHA,
		AppliedSHA256:  sha256Hex(candidate),
		BackupID:       backupID,
		MainPID:        after.mainPID,
		NRestarts:      after.nRestarts,
	}, nil
}

func (o *FixedOperator) Rollback(ctx context.Context, request RollbackRequest) (RollbackResponse, error) {
	if !validBackupID(request.BackupID) || !strings.HasPrefix(request.BackupID, "backup-") {
		return RollbackResponse{}, errors.New("runtimeagent: invalid backup id")
	}
	backupPath, err := o.backupPath(request.BackupID)
	if err != nil {
		return RollbackResponse{}, err
	}
	restored, err := os.ReadFile(backupPath)
	if err != nil {
		return RollbackResponse{}, fmt.Errorf("runtimeagent: read rollback backup: %w", err)
	}
	if _, err := naiveruntime.InspectCaddyfile(restored); err != nil {
		return RollbackResponse{}, fmt.Errorf("runtimeagent: rollback backup is not a supported Caddyfile: %w", err)
	}

	current, metadata, err := o.readCurrent()
	if err != nil {
		return RollbackResponse{}, err
	}
	candidatePath, err := o.writeTemp(restored, metadata)
	if err != nil {
		return RollbackResponse{}, err
	}
	defer os.Remove(candidatePath)
	if err := o.validatePath(ctx, candidatePath); err != nil {
		return RollbackResponse{}, err
	}
	before, err := o.snapshotService(ctx)
	if err != nil {
		return RollbackResponse{}, err
	}
	if err := o.installTemp(candidatePath); err != nil {
		return RollbackResponse{}, err
	}
	after, err := o.reloadAndVerify(ctx, before)
	if err != nil {
		compensationErr := o.restoreAndReload(ctx, current, metadata, before)
		if compensationErr != nil {
			return RollbackResponse{}, fmt.Errorf("runtimeagent: rollback failed and current-state compensation failed: %v; compensation: %w", err, compensationErr)
		}
		return RollbackResponse{}, fmt.Errorf("runtimeagent: rollback failed and current state was restored: %w", err)
	}
	return RollbackResponse{
		RestoredSHA256: sha256Hex(restored),
		MainPID:        after.mainPID,
		NRestarts:      after.nRestarts,
	}, nil
}

func (o *FixedOperator) prepareCandidate(expectedSHA string, desired DesiredStateInput) ([]byte, fileMetadata, error) {
	current, metadata, err := o.readCurrent()
	if err != nil {
		return nil, fileMetadata{}, err
	}
	if expectedSHA != sha256Hex(current) {
		return nil, fileMetadata{}, errors.New("runtimeagent: current Caddyfile SHA does not match expected revision")
	}
	if !validMutationInput(expectedSHA, desired) {
		return nil, fileMetadata{}, errors.New("runtimeagent: invalid desired state")
	}
	candidate, err := renderDesired(current, desired)
	if err != nil {
		return nil, fileMetadata{}, err
	}
	return candidate, metadata, nil
}

func renderDesired(current []byte, desired DesiredStateInput) ([]byte, error) {
	credentials := make([]runtimecred.DesiredCredential, 0, len(desired.Credentials))
	for _, input := range desired.Credentials {
		credential, err := runtimecred.NewImportedDesiredCredential(input.ID, input.Username, input.Password, input.Status)
		if err != nil {
			return nil, fmt.Errorf("runtimeagent: invalid desired credential: %w", err)
		}
		credentials = append(credentials, credential)
	}
	candidate, err := naiveruntime.RenderCredentials(current, credentials)
	if err != nil {
		return nil, fmt.Errorf("runtimeagent: render Caddy credentials: %w", err)
	}
	return candidate, nil
}

func (o *FixedOperator) readCurrent() ([]byte, fileMetadata, error) {
	info, err := os.Lstat(o.config.caddyfilePath)
	if err != nil {
		return nil, fileMetadata{}, fmt.Errorf("runtimeagent: stat Caddyfile: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fileMetadata{}, errors.New("runtimeagent: Caddyfile is not a direct regular file")
	}
	content, err := os.ReadFile(o.config.caddyfilePath)
	if err != nil {
		return nil, fileMetadata{}, fmt.Errorf("runtimeagent: read Caddyfile: %w", err)
	}
	metadata := fileMetadata{mode: info.Mode().Perm(), uid: os.Geteuid(), gid: os.Getegid()}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		metadata.uid = int(stat.Uid)
		metadata.gid = int(stat.Gid)
	}
	return content, metadata, nil
}

func (o *FixedOperator) writeTemp(content []byte, metadata fileMetadata) (string, error) {
	directory := filepath.Dir(o.config.caddyfilePath)
	file, err := os.CreateTemp(directory, ".pvnaive-candidate-*")
	if err != nil {
		return "", fmt.Errorf("runtimeagent: create candidate: %w", err)
	}
	path := file.Name()
	clean := func(cause error) (string, error) {
		_ = file.Close()
		_ = os.Remove(path)
		return "", cause
	}
	if err := file.Chmod(metadata.mode.Perm()); err != nil {
		return clean(fmt.Errorf("runtimeagent: chmod candidate: %w", err))
	}
	if os.Geteuid() == 0 {
		if err := file.Chown(metadata.uid, metadata.gid); err != nil {
			return clean(fmt.Errorf("runtimeagent: chown candidate: %w", err))
		}
	}
	if _, err := file.Write(content); err != nil {
		return clean(fmt.Errorf("runtimeagent: write candidate: %w", err))
	}
	if err := file.Sync(); err != nil {
		return clean(fmt.Errorf("runtimeagent: fsync candidate: %w", err))
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("runtimeagent: close candidate: %w", err)
	}
	return path, nil
}

func (o *FixedOperator) installTemp(path string) error {
	if filepath.Dir(path) != filepath.Dir(o.config.caddyfilePath) {
		return errors.New("runtimeagent: candidate is not on Caddyfile filesystem")
	}
	if err := os.Rename(path, o.config.caddyfilePath); err != nil {
		return fmt.Errorf("runtimeagent: atomically install Caddyfile: %w", err)
	}
	return syncDirectory(filepath.Dir(o.config.caddyfilePath))
}

func (o *FixedOperator) validatePath(ctx context.Context, path string) error {
	commandCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	if _, err := o.runner.Run(commandCtx, o.config.caddyBinary, "validate", "--config", path, "--adapter", "caddyfile"); err != nil {
		return errors.New("runtimeagent: Caddy candidate validation failed")
	}
	return nil
}

func (o *FixedOperator) createBackup(content []byte) (string, error) {
	if err := os.MkdirAll(o.config.backupRoot, 0700); err != nil {
		return "", fmt.Errorf("runtimeagent: create backup root: %w", err)
	}
	for attempt := 0; attempt < 8; attempt++ {
		randomBytes := make([]byte, 4)
		if _, err := io.ReadFull(o.random, randomBytes); err != nil {
			return "", fmt.Errorf("runtimeagent: generate backup id: %w", err)
		}
		backupID := "backup-" + o.now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(randomBytes)
		directory := filepath.Join(o.config.backupRoot, backupID)
		if err := os.Mkdir(directory, 0700); err != nil {
			if errors.Is(err, os.ErrExist) {
				continue
			}
			return "", fmt.Errorf("runtimeagent: create backup directory: %w", err)
		}
		path := filepath.Join(directory, "Caddyfile")
		if err := writeSyncedFile(path, content, 0600); err != nil {
			_ = os.RemoveAll(directory)
			return "", err
		}
		checksum := []byte(sha256Hex(content) + "  Caddyfile\n")
		if err := writeSyncedFile(filepath.Join(directory, "Caddyfile.sha256"), checksum, 0600); err != nil {
			_ = os.RemoveAll(directory)
			return "", err
		}
		if err := syncDirectory(directory); err != nil {
			return "", err
		}
		if err := syncDirectory(o.config.backupRoot); err != nil {
			return "", err
		}
		return backupID, nil
	}
	return "", errors.New("runtimeagent: unable to allocate unique backup id")
}

func (o *FixedOperator) backupPath(backupID string) (string, error) {
	candidate := filepath.Join(o.config.backupRoot, backupID, "Caddyfile")
	relative, err := filepath.Rel(o.config.backupRoot, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", errors.New("runtimeagent: backup path escaped fixed root")
	}
	return candidate, nil
}

func (o *FixedOperator) snapshotService(ctx context.Context) (serviceSnapshot, error) {
	active, err := o.runSystemctl(ctx, "is-active", o.config.serviceName)
	if err != nil || strings.TrimSpace(string(active)) != "active" {
		return serviceSnapshot{}, errors.New("runtimeagent: Caddy service is not active")
	}
	pidRaw, err := o.runSystemctl(ctx, "show", "--property=MainPID", "--value", o.config.serviceName)
	if err != nil {
		return serviceSnapshot{}, errors.New("runtimeagent: cannot read Caddy MainPID")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidRaw)))
	if err != nil || pid <= 0 {
		return serviceSnapshot{}, errors.New("runtimeagent: invalid Caddy MainPID")
	}
	restartsRaw, err := o.runSystemctl(ctx, "show", "--property=NRestarts", "--value", o.config.serviceName)
	if err != nil {
		return serviceSnapshot{}, errors.New("runtimeagent: cannot read Caddy NRestarts")
	}
	restarts, err := strconv.Atoi(strings.TrimSpace(string(restartsRaw)))
	if err != nil || restarts < 0 {
		return serviceSnapshot{}, errors.New("runtimeagent: invalid Caddy NRestarts")
	}
	return serviceSnapshot{mainPID: pid, nRestarts: restarts}, nil
}

func (o *FixedOperator) reloadAndVerify(ctx context.Context, before serviceSnapshot) (serviceSnapshot, error) {
	if _, err := o.runSystemctl(ctx, "reload", o.config.serviceName); err != nil {
		return serviceSnapshot{}, errors.New("runtimeagent: Caddy reload failed")
	}
	after, err := o.snapshotService(ctx)
	if err != nil {
		return serviceSnapshot{}, err
	}
	if after.mainPID != before.mainPID {
		return serviceSnapshot{}, errors.New("runtimeagent: Caddy MainPID changed during reload")
	}
	if after.nRestarts != before.nRestarts {
		return serviceSnapshot{}, errors.New("runtimeagent: Caddy restart counter changed during reload")
	}
	return after, nil
}

func (o *FixedOperator) runSystemctl(ctx context.Context, args ...string) ([]byte, error) {
	commandCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	return o.runner.Run(commandCtx, "systemctl", args...)
}

func (o *FixedOperator) restoreAndReload(ctx context.Context, content []byte, metadata fileMetadata, expected serviceSnapshot) error {
	path, err := o.writeTemp(content, metadata)
	if err != nil {
		return err
	}
	defer os.Remove(path)
	if err := o.validatePath(ctx, path); err != nil {
		return fmt.Errorf("runtimeagent: validate compensation Caddyfile: %w", err)
	}
	if err := o.installTemp(path); err != nil {
		return err
	}
	if _, err := o.reloadAndVerify(ctx, expected); err != nil {
		return err
	}
	installed, _, err := o.readCurrent()
	if err != nil {
		return err
	}
	if sha256Hex(installed) != sha256Hex(content) {
		return errors.New("runtimeagent: compensation bytes did not restore exactly")
	}
	return nil
}

func writeSyncedFile(path string, content []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("runtimeagent: create backup file: %w", err)
	}
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("runtimeagent: write backup file: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("runtimeagent: fsync backup file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("runtimeagent: close backup file: %w", err)
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("runtimeagent: open directory for fsync: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("runtimeagent: fsync directory: %w", err)
	}
	return nil
}

func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
