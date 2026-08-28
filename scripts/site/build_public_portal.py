#!/usr/bin/env python3
from __future__ import annotations

import argparse
import html
import json
import re
from pathlib import Path
from urllib.parse import urlparse

REPO_ROOT = Path(__file__).resolve().parents[2]
SITE_ROOT = REPO_ROOT / "site"
DATA_ROOT = SITE_ROOT / "data"

ALLOWED_MEDIA_MIME = {
    "video/mp4",
    "video/webm",
    "audio/mpeg",
    "audio/mp4",
    "audio/ogg",
    "application/pdf",
}


def load_json(path: Path):
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def validate_external_url(value: str) -> str:
    parsed = urlparse(value)
    if parsed.scheme != "https" or not parsed.netloc:
        raise ValueError(f"external URL must be HTTPS: {value}")
    return value


def validate_local_path(value: str) -> str:
    if not value or value.startswith("/") or "\\" in value:
        raise ValueError(f"invalid local path: {value}")
    parts = Path(value).parts
    if ".." in parts or "." in parts:
        raise ValueError(f"invalid local path: {value}")
    if not re.fullmatch(r"[A-Za-z0-9._/-]+", value):
        raise ValueError(f"unsafe local path: {value}")
    return value


def safe_url(value: str) -> str:
    if value.startswith("/"):
        if ".." in value or "\\" in value:
            raise ValueError(f"unsafe local URL: {value}")
        return value
    return validate_external_url(value)


def human_bytes(value):
    if value in (None, ""):
        return None
    value = int(value)
    units = ["B", "KB", "MB", "GB"]
    n = float(value)
    for unit in units:
        if n < 1024 or unit == units[-1]:
            if unit == "B":
                return f"{int(n):,} {unit}"
            return f"{n:.2f} {unit}"
        n /= 1024
    return f"{value:,} B"


def load_and_validate_media(path: Path = DATA_ROOT / "media.json"):
    payload = load_json(path)
    items = payload.get("items")
    if not isinstance(items, list):
        raise ValueError("media items must be a list")
    ids, slugs = set(), set()
    for item in items:
        required = ["id", "slug", "kind", "title", "summary", "source_name", "source_domain", "source_url", "qualities"]
        for key in required:
            if key not in item or item[key] in (None, ""):
                raise ValueError(f"media item missing {key}")
        if item["id"] in ids or item["slug"] in slugs:
            raise ValueError("duplicate media id/slug")
        ids.add(item["id"])
        slugs.add(item["slug"])
        validate_external_url(item["source_url"])
        if item.get("poster"):
            safe_url(item["poster"])
        if item.get("kind") not in {"video", "audio", "pdf", "document"}:
            raise ValueError(f"unsupported media kind: {item.get('kind')}")
        if item.get("mirror_allowed") and not str(item.get("rights_note", "")).strip():
            raise ValueError("mirror_allowed item requires rights_note")
        for quality in item["qualities"]:
            mime = quality.get("mime")
            if mime not in ALLOWED_MEDIA_MIME:
                raise ValueError(f"unsupported media MIME: {mime}")
            validate_external_url(quality.get("url", ""))
            local_path = quality.get("local_path")
            if local_path is not None:
                validate_local_path(local_path)
            if item.get("mirror_allowed"):
                if not local_path:
                    raise ValueError("mirrored quality requires local_path")
                if not isinstance(quality.get("bytes"), int) or quality["bytes"] <= 0:
                    raise ValueError("mirrored quality requires verified positive byte size")
            if quality.get("sha256") is not None and not re.fullmatch(r"[0-9a-f]{64}", quality["sha256"]):
                raise ValueError("invalid sha256")
    return payload


def load_portal():
    portal = load_json(DATA_ROOT / "portal.json")
    for route in portal.get("routes", []):
        path = route.get("path", "")
        if not (path.startswith("/") and path.endswith("/") and ".." not in path):
            raise ValueError(f"invalid route path: {path}")
    return portal


def load_articles():
    payload = load_json(DATA_ROOT / "articles.json")
    for source in payload.get("sources", []):
        validate_external_url(source["homepage"])
    for article in payload.get("articles", []):
        validate_external_url(article["source_url"])
        if article.get("image"):
            safe_url(article["image"])
    return payload


def load_galleries():
    payload = load_json(DATA_ROOT / "galleries.json")
    seen = set()
    for gallery in payload.get("galleries", []):
        slug = gallery.get("slug")
        if not slug or slug in seen:
            raise ValueError("gallery slug missing/duplicate")
        seen.add(slug)
        validate_external_url(gallery["source_url"])
        for image in gallery.get("images", []):
            safe_url(image["src"])
    return payload


def esc(value) -> str:
    return html.escape(str(value or ""), quote=True)


def link(href: str, label: str, css: str = "") -> str:
    href = safe_url(href)
    attrs = f' class="{esc(css)}"' if css else ""
    external = href.startswith("https://")
    target = ' target="_blank" rel="noopener noreferrer"' if external else ""
    return f'<a href="{esc(href)}"{attrs}{target}>{esc(label)}</a>'


def nav_html(portal) -> str:
    return "".join(link(route["path"], route["label"]) for route in portal["routes"])


def shell(portal, title: str, body: str, breadcrumb: str = "") -> str:
    disclosure = esc(portal["disclosure"])
    return f'''<!doctype html>
<html lang="fa" dir="rtl">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <meta name="theme-color" content="#101b27">
  <meta name="description" content="{disclosure}">
  <link rel="stylesheet" href="/assets/site.css">
  <link rel="icon" href="/assets/news-mark.svg" type="image/svg+xml">
  <title>{esc(title)} | {esc(portal['name'])}</title>
</head>
<body class="portal-page">
  <header class="portal-masthead">
    <div class="shell portal-masthead__inner">
      <a class="wordmark" href="/"><img src="/assets/news-mark.svg" width="44" height="44" alt=""><span><strong>{esc(portal['name'])}</strong><small>{esc(portal['tagline'])}</small></span></a>
      <nav class="portal-nav" aria-label="ناوبری اصلی">{nav_html(portal)}</nav>
    </div>
  </header>
  <main class="shell portal-main">
    <div class="portal-breadcrumb"><a href="/">خانه</a>{breadcrumb}</div>
    {body}
  </main>
  <footer class="site-footer"><div class="shell footer-grid"><section><strong>{esc(portal['name'])}</strong><p>{disclosure}</p></section><section><h2>دسترسی سریع</h2>{nav_html(portal)}</section><section><h2>سیاست منبع</h2><p>هر مطلب دارای پیوند مستقیم به منبع اصلی است. متن کامل آثار دارای حق نشر بازنشر نمی‌شود.</p></section></div></footer>
</body>
</html>
'''


def write_page(root: Path, relative: str, content: str):
    target = root / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(content, encoding="utf-8", newline="\n")


def card(title: str, summary: str, href: str, meta: str = "") -> str:
    return f'<article class="portal-card"><div class="story-meta"><span>{esc(meta)}</span></div><h2>{link(href, title)}</h2><p>{esc(summary)}</p></article>'


def build_news(portal, articles, out_root: Path):
    cards = []
    for article in articles.get("articles", []):
        slug = article["id"]
        local = f"/news/{slug}.html"
        cards.append(card(article["title"], article["summary"], local, article.get("published_label", "")))
        body = f'''<article class="portal-detail"><div class="story-meta"><span>{esc(article.get('category'))}</span><span>{esc(article.get('published_label'))}</span></div><h1>{esc(article['title'])}</h1><p class="portal-lead">{esc(article['summary'])}</p><div class="portal-source-box"><strong>منبع اصلی</strong>{link(article['source_url'], 'مشاهده مطلب در منبع اصلی')}</div><div class="portal-related"><h2>ادامه مرور</h2>{link('/news/', 'بازگشت به آرشیو خبرها')} · {link('/sources/', 'فهرست منابع')}</div></article>'''
        write_page(out_root, f"news/{slug}.html", shell(portal, article["title"], body, ' / <a href="/news/">خبرها</a>'))
    body = f'<section class="portal-heading"><span>آرشیو</span><h1>خبرها و ارجاعات</h1><p>مرور خلاصه‌های منبع‌دار؛ برای متن کامل به منبع اصلی بروید.</p></section><div class="portal-grid">{"".join(cards)}</div>'
    write_page(out_root, "news/index.html", shell(portal, "خبرها", body, ' / خبرها'))


def player_sources(item):
    sources = []
    for quality in item["qualities"]:
        if item.get("mirror_allowed") and quality.get("local_path"):
            sources.append((f"/media/{quality['local_path']}", quality["mime"]))
        sources.append((quality["url"], quality["mime"]))
    return sources


def build_media_kind(portal, media, out_root: Path, kind: str):
    dirname = "videos" if kind == "video" else "audio"
    label = "ویدئو" if kind == "video" else "صوت"
    items = [item for item in media["items"] if item["kind"] == kind]
    cards = []
    for item in items:
        local = f"/{dirname}/{item['slug']}.html"
        cards.append(card(item["title"], item["summary"], local, item.get("published_label", "")))
        tag = "video" if kind == "video" else "audio"
        attrs = ' controls preload="metadata"'
        if kind == "video" and item.get("poster"):
            attrs += f' poster="{esc(safe_url(item["poster"]))}"'
        src_html = "".join(f'<source src="{esc(safe_url(src))}" type="{esc(mime)}">' for src, mime in player_sources(item))
        rows = []
        for quality in item["qualities"]:
            size = human_bytes(quality.get("bytes")) or "اندازه در منبع تثبیت نشده"
            if item.get("mirror_allowed") and quality.get("local_path"):
                local_dl = f'/media/{quality["local_path"]}'
                receive = f'<a href="{esc(local_dl)}" download>دانلود از این سایت</a> · {link(quality["url"], "منبع مستقیم")}'
            else:
                receive = link(quality["url"], "پخش/دانلود از منبع")
            rows.append(f'<tr><td>{esc(quality["label"])}</td><td>{esc(quality["mime"])}</td><td>{esc(size)}</td><td>{receive}</td></tr>')
        body = f'''<article class="portal-detail media-detail"><div class="story-meta"><span>{esc(label)}</span><span>{esc(item.get('published_label'))}</span></div><h1>{esc(item['title'])}</h1><p class="portal-lead">{esc(item['summary'])}</p><div class="media-player"><{tag}{attrs}>{src_html}مرورگر شما پخش این فایل را پشتیبانی نمی‌کند.</{tag}></div><div class="portal-source-box"><strong>انتساب و مجوز</strong><p>{esc(item.get('attribution'))}</p><p>{esc(item.get('rights_note'))}</p>{link(item['source_url'], 'صفحه منبع')}</div><div class="download-table"><table><thead><tr><th>کیفیت</th><th>فرمت</th><th>حجم</th><th>دریافت</th></tr></thead><tbody>{''.join(rows)}</tbody></table></div><div class="portal-related">{link(f'/{dirname}/', f'بازگشت به آرشیو {label}')} · {link('/downloads/', 'همه دانلودها')} · {link('/sources/', 'منابع')}</div></article>'''
        write_page(out_root, f"{dirname}/{item['slug']}.html", shell(portal, item["title"], body, f' / <a href="/{dirname}/">{label}</a>'))
    body = f'<section class="portal-heading"><span>چندرسانه‌ای</span><h1>آرشیو {esc(label)}</h1><p>پخش آنلاین و دسترسی به فایل‌های منبع‌دار.</p></section><div class="portal-grid">{"".join(cards)}</div>'
    write_page(out_root, f"{dirname}/index.html", shell(portal, f"آرشیو {label}", body, f' / {label}'))


def build_galleries(portal, galleries, out_root: Path):
    cards = []
    for gallery in galleries.get("galleries", []):
        local = f"/gallery/{gallery['slug']}.html"
        cards.append(card(gallery["title"], gallery["summary"], local, gallery.get("published_label", "")))
        images = "".join(f'<figure><img src="{esc(safe_url(image["src"]))}" loading="lazy" alt="{esc(image["alt"])}"><figcaption>{esc(image.get("caption"))}</figcaption></figure>' for image in gallery.get("images", []))
        body = f'<article class="portal-detail"><h1>{esc(gallery["title"])}</h1><p class="portal-lead">{esc(gallery["summary"])}</p><div class="portal-gallery">{images}</div><div class="portal-source-box">{link(gallery["source_url"], "مشاهده منبع")}</div><div class="portal-related">{link("/gallery/", "بازگشت به گالری‌ها")}</div></article>'
        write_page(out_root, f"gallery/{gallery['slug']}.html", shell(portal, gallery["title"], body, ' / <a href="/gallery/">گالری</a>'))
    body = f'<section class="portal-heading"><span>تصویر</span><h1>گالری‌ها</h1></section><div class="portal-grid">{"".join(cards)}</div>'
    write_page(out_root, "gallery/index.html", shell(portal, "گالری", body, ' / گالری'))


def build_downloads(portal, media, out_root: Path):
    rows = []
    for item in media["items"]:
        detail_dir = "videos" if item["kind"] == "video" else "audio"
        detail_href = f"/{detail_dir}/{item['slug']}.html"
        for quality in item["qualities"]:
            size = human_bytes(quality.get("bytes")) or "—"
            if item.get("mirror_allowed") and quality.get("local_path"):
                local_dl = f'/media/{quality["local_path"]}'
                receive = f'<a href="{esc(local_dl)}" download>دانلود محلی</a> · {link(quality["url"], "نسخه منبع")}'
            else:
                receive = link(quality["url"], "دانلود از منبع")
            title_link = link(detail_href, item["title"])
            rows.append(f'<tr><td>{title_link}</td><td>{esc(quality["mime"])}</td><td>{esc(size)}</td><td>{esc(item["source_name"])}</td><td>{receive}</td></tr>')
    body = f'<section class="portal-heading"><span>کتابخانه فایل</span><h1>دانلودها</h1><p>حجم فقط زمانی نمایش داده می‌شود که مقدار آن در manifest تأیید شده باشد.</p></section><div class="download-table"><table><thead><tr><th>عنوان</th><th>فرمت</th><th>حجم</th><th>منبع</th><th>دریافت</th></tr></thead><tbody>{"".join(rows)}</tbody></table></div>'
    write_page(out_root, "downloads/index.html", shell(portal, "دانلودها", body, ' / دانلودها'))


def build_sources(portal, articles, media, out_root: Path):
    sources = {}
    for source in articles.get("sources", []):
        sources[source["domain"]] = (source["name"], source["homepage"])
    for item in media["items"]:
        sources.setdefault(item["source_domain"], (item["source_name"], item["source_url"]))
    cards = [card(name, domain, url, domain) for domain, (name, url) in sorted(sources.items())]
    body = f'<section class="portal-heading"><span>شفافیت</span><h1>فهرست منابع</h1><p>هر منبع با نام و پیوند مستقیم نمایش داده می‌شود.</p></section><div class="portal-grid">{"".join(cards)}</div>'
    write_page(out_root, "sources/index.html", shell(portal, "منابع", body, ' / منابع'))


def build_about(portal, out_root: Path):
    body = f'<article class="portal-detail"><h1>درباره این درگاه</h1><p class="portal-lead">{esc(portal["disclosure"])}</p><h2>سیاست انتشار</h2><p>خلاصه‌ها و فراداده‌ها برای هدایت کاربر به منبع اصلی ارائه می‌شوند. فایل محلی فقط زمانی نگهداری می‌شود که مجوز بازنشر روشن باشد و انتساب و شرایط مجوز همراه آن باقی بماند.</p><h2>پایداری</h2><p>صفحات به‌صورت استاتیک ساخته می‌شوند تا اختلال منبع خارجی، ساختار اصلی سایت را از دسترس خارج نکند.</p></article>'
    write_page(out_root, "about/index.html", shell(portal, "درباره", body, ' / درباره'))


def build_all(out_root: Path):
    portal = load_portal()
    articles = load_articles()
    media = load_and_validate_media()
    galleries = load_galleries()
    build_news(portal, articles, out_root)
    build_media_kind(portal, media, out_root, "video")
    build_media_kind(portal, media, out_root, "audio")
    build_galleries(portal, galleries, out_root)
    build_downloads(portal, media, out_root)
    build_sources(portal, articles, media, out_root)
    build_about(portal, out_root)


def validate_all():
    load_portal()
    load_articles()
    load_and_validate_media()
    load_galleries()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    parser.add_argument("--output-root", type=Path, default=SITE_ROOT)
    args = parser.parse_args()
    validate_all()
    if args.check:
        print("PUBLIC_PORTAL_MANIFEST_CHECK=PASSED")
        return
    build_all(args.output_root)
    print(f"PUBLIC_PORTAL_BUILD=PASSED output={args.output_root}")


if __name__ == "__main__":
    main()
