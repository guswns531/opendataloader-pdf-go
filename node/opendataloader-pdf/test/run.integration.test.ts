import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { run, convert } from '../src/index';
import * as path from 'path';
import * as fs from 'fs';
import { execFileSync } from 'child_process';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const rootDir = path.resolve(__dirname, '..', '..', '..');
const inputPdf = path.join(rootDir, 'samples', 'pdf', '1901.03003.pdf');
const tempDir = path.join(__dirname, 'temp', 'run');
const goBin = path.join(tempDir, 'opendataloader-pdf');

describe('opendataloader-pdf', () => {
  beforeAll(() => {
    // Clean up previous test runs
    if (fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true, force: true });
    }
    fs.mkdirSync(tempDir, { recursive: true });
    execFileSync('go', ['build', '-o', goBin, './cmd/opendataloader-pdf'], { cwd: rootDir });
    process.env.OPENDATALOADER_PDF_CLI_BIN = goBin;
  });

  afterAll(() => {
    delete process.env.OPENDATALOADER_PDF_CLI_BIN;
    // Clean up after tests
    if (fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true, force: true });
    }
  });

  it('should process PDF and generate markdown output', async () => {
    console.log(`[TEST] Running opendataloader-pdf test...`);
    console.log(`[TEST] Input PDF: ${inputPdf}`);
    console.log(`[TEST] Output directory: ${tempDir}`);

    await run(inputPdf, {
      outputFolder: tempDir,
      generateMarkdown: true,
      generateHtml: true,
      debug: true,
    });

    expect(fs.existsSync(path.join(tempDir, '1901.03003.json'))).toBe(true);
    expect(fs.existsSync(path.join(tempDir, '1901.03003.md'))).toBe(true);
    expect(fs.existsSync(path.join(tempDir, '1901.03003.html'))).toBe(true);
  }, 30000); // 30 second timeout for this test

  it('should convert PDF with explicit formats using quiet mode', async () => {
    const convertDir = path.join(tempDir, 'convert');
    if (fs.existsSync(convertDir)) {
      fs.rmSync(convertDir, { recursive: true, force: true });
    }
    fs.mkdirSync(convertDir);

    await convert([inputPdf], {
      outputDir: convertDir,
      format: ['json', 'text', 'html', 'markdown'],
    });

    expect(fs.existsSync(path.join(convertDir, '1901.03003.json'))).toBe(true);
    expect(fs.existsSync(path.join(convertDir, '1901.03003.txt'))).toBe(true);
    expect(fs.existsSync(path.join(convertDir, '1901.03003.html'))).toBe(true);
    expect(fs.existsSync(path.join(convertDir, '1901.03003.md'))).toBe(true);
  }, 30000);
});
