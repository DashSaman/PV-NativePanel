const VERSION = 6;
const SIZE = 41;
const DATA_CODEWORDS = 136;
const DATA_BYTES_PER_BLOCK = 68;
const ECC_CODEWORDS_PER_BLOCK = 18;
const MAX_BYTE_LENGTH = 134;

function appendBits(bits: boolean[], value: number, length: number) {
  for (let i = length - 1; i >= 0; i -= 1) bits.push(((value >>> i) & 1) !== 0);
}

function multiplyGF(x: number, y: number): number {
  let z = 0;
  for (let i = 7; i >= 0; i -= 1) {
    z = (z << 1) ^ ((z >>> 7) * 0x11d);
    z ^= ((y >>> i) & 1) * x;
  }
  return z;
}

function computeDivisor(degree: number): Uint8Array {
  const result = new Uint8Array(degree);
  result[degree - 1] = 1;
  let root = 1;
  for (let i = 0; i < degree; i += 1) {
    for (let j = 0; j < degree; j += 1) {
      result[j] = multiplyGF(result[j], root);
      if (j + 1 < degree) result[j] ^= result[j + 1];
    }
    root = multiplyGF(root, 0x02);
  }
  return result;
}

function computeRemainder(data: Uint8Array, divisor: Uint8Array): Uint8Array {
  const result = new Uint8Array(divisor.length);
  for (const value of data) {
    const factor = value ^ result[0];
    result.copyWithin(0, 1);
    result[result.length - 1] = 0;
    for (let i = 0; i < divisor.length; i += 1) {
      result[i] ^= multiplyGF(divisor[i], factor);
    }
  }
  return result;
}

function buildDataCodewords(text: string): Uint8Array {
  const payload = new TextEncoder().encode(text);
  if (payload.length > MAX_BYTE_LENGTH) throw new Error("QR payload is too long for local Version 6-L encoding");

  const bits: boolean[] = [];
  appendBits(bits, 0b0100, 4);
  appendBits(bits, payload.length, 8);
  for (const value of payload) appendBits(bits, value, 8);

  const capacity = DATA_CODEWORDS * 8;
  const terminator = Math.min(4, capacity - bits.length);
  for (let i = 0; i < terminator; i += 1) bits.push(false);
  while (bits.length % 8 !== 0) bits.push(false);

  const data: number[] = [];
  for (let i = 0; i < bits.length; i += 8) {
    let value = 0;
    for (let j = 0; j < 8; j += 1) value = (value << 1) | (bits[i + j] ? 1 : 0);
    data.push(value);
  }
  for (let pad = 0; data.length < DATA_CODEWORDS; pad += 1) data.push(pad % 2 === 0 ? 0xec : 0x11);
  return Uint8Array.from(data);
}

function buildCodewords(text: string): Uint8Array {
  const data = buildDataCodewords(text);
  const divisor = computeDivisor(ECC_CODEWORDS_PER_BLOCK);
  const blocks = [
    data.slice(0, DATA_BYTES_PER_BLOCK),
    data.slice(DATA_BYTES_PER_BLOCK, DATA_CODEWORDS),
  ];
  const ecc = blocks.map((block) => computeRemainder(block, divisor));
  const output: number[] = [];
  for (let i = 0; i < DATA_BYTES_PER_BLOCK; i += 1) {
    output.push(blocks[0][i], blocks[1][i]);
  }
  for (let i = 0; i < ECC_CODEWORDS_PER_BLOCK; i += 1) {
    output.push(ecc[0][i], ecc[1][i]);
  }
  return Uint8Array.from(output);
}

function formatBits(mask: number): number {
  const data = (1 << 3) | mask; // Error-correction level L has format bits 01.
  let remainder = data;
  for (let i = 0; i < 10; i += 1) {
    remainder = (remainder << 1) ^ (((remainder >>> 9) & 1) * 0x537);
  }
  return ((data << 10) | remainder) ^ 0x5412;
}

export function encodeQR(text: string): boolean[][] {
  if (VERSION !== 6) throw new Error("unsupported QR version");
  const modules = Array.from({ length: SIZE }, () => Array<boolean>(SIZE).fill(false));
  const functions = Array.from({ length: SIZE }, () => Array<boolean>(SIZE).fill(false));

  const setFunction = (x: number, y: number, dark: boolean) => {
    if (x < 0 || y < 0 || x >= SIZE || y >= SIZE) return;
    modules[y][x] = dark;
    functions[y][x] = true;
  };

  for (let i = 8; i < SIZE - 8; i += 1) {
    setFunction(6, i, i % 2 === 0);
    setFunction(i, 6, i % 2 === 0);
  }

  for (const [centerX, centerY] of [[3, 3], [SIZE - 4, 3], [3, SIZE - 4]]) {
    for (let dy = -4; dy <= 4; dy += 1) {
      for (let dx = -4; dx <= 4; dx += 1) {
        const distance = Math.max(Math.abs(dx), Math.abs(dy));
        setFunction(centerX + dx, centerY + dy, distance !== 2 && distance !== 4);
      }
    }
  }

  // Version 6 has alignment centers [6, 34]. Three overlap finder patterns;
  // only the bottom-right alignment pattern is drawn.
  for (let dy = -2; dy <= 2; dy += 1) {
    for (let dx = -2; dx <= 2; dx += 1) {
      setFunction(34 + dx, 34 + dy, Math.max(Math.abs(dx), Math.abs(dy)) !== 1);
    }
  }

  const format = formatBits(0);
  const bit = (index: number) => ((format >>> index) & 1) !== 0;
  for (let i = 0; i <= 5; i += 1) setFunction(8, i, bit(i));
  setFunction(8, 7, bit(6));
  setFunction(8, 8, bit(7));
  setFunction(7, 8, bit(8));
  for (let i = 9; i < 15; i += 1) setFunction(14 - i, 8, bit(i));
  for (let i = 0; i < 8; i += 1) setFunction(SIZE - 1 - i, 8, bit(i));
  for (let i = 8; i < 15; i += 1) setFunction(8, SIZE - 15 + i, bit(i));
  setFunction(8, SIZE - 8, true);

  const codewords = buildCodewords(text);
  let bitIndex = 0;
  for (let right = SIZE - 1; right >= 1; right -= 2) {
    if (right === 6) right = 5;
    const upward = ((right + 1) & 2) === 0;
    for (let vertical = 0; vertical < SIZE; vertical += 1) {
      const y = upward ? SIZE - 1 - vertical : vertical;
      for (let j = 0; j < 2; j += 1) {
        const x = right - j;
        if (functions[y][x] || bitIndex >= codewords.length * 8) continue;
        modules[y][x] = ((codewords[bitIndex >>> 3] >>> (7 - (bitIndex & 7))) & 1) !== 0;
        bitIndex += 1;
      }
    }
  }

  if (bitIndex !== codewords.length * 8) throw new Error("QR data placement failed");

  // Fixed mask 0. Apply it to every non-function module, including remainder bits.
  for (let y = 0; y < SIZE; y += 1) {
    for (let x = 0; x < SIZE; x += 1) {
      if (!functions[y][x] && (x + y) % 2 === 0) modules[y][x] = !modules[y][x];
    }
  }
  return modules;
}
