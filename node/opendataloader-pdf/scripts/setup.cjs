const fs = require('fs');
const path = require('path');

const rootDir = path.resolve(__dirname, '..');
const binaryName = process.platform === 'win32' ? 'opendataloader-pdf.exe' : 'opendataloader-pdf';
const envBinary = process.env.OPENDATALOADER_PDF_BIN;
const candidateBinaries = [
  envBinary,
  path.join(rootDir, '../../bin', binaryName),
  path.join(rootDir, '../../go', binaryName),
].filter(Boolean);

const sourceBinaryPath = candidateBinaries.find((candidate) => fs.existsSync(candidate));
if (!sourceBinaryPath) {
  console.error(
    'Could not find the opendataloader-pdf binary. ' +
      "Set OPENDATALOADER_PDF_BIN or run 'cd go && go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf/' first.",
  );
  console.error(`Searched: ${candidateBinaries.join(', ')}`);
  process.exit(1);
}

console.log(`Found source binary: ${sourceBinaryPath}`);

const destBinDir = path.join(rootDir, 'bin').replace(/\\/g, '/');
if (!fs.existsSync(destBinDir)) {
  fs.mkdirSync(destBinDir, { recursive: true });
}
const destBinaryPath = path.join(destBinDir, binaryName).replace(/\\/g, '/');
console.log(`Copying binary to ${destBinaryPath}`);
fs.copyFileSync(sourceBinaryPath, destBinaryPath);
if (process.platform !== 'win32') {
  fs.chmodSync(destBinaryPath, 0o755);
}

// Copy README.md, LICENSE, NOTICE, and THIRD_PARTY
const readmeSrc = path.resolve(rootDir, '../../README.md');
const licenseSrc = path.resolve(rootDir, '../../LICENSE');
const noticeSrc = path.resolve(rootDir, '../../NOTICE');
const thirdPartySrc = path.resolve(rootDir, '../../THIRD_PARTY');

const readmeDest = path.join(rootDir, 'README.md').replace(/\\/g, '/');
const licenseDest = path.join(rootDir, 'LICENSE').replace(/\\/g, '/');
const noticeDest = path.join(rootDir, 'NOTICE').replace(/\\/g, '/');
const thirdPartyDest = path.join(rootDir, 'THIRD_PARTY').replace(/\\/g, '/');

console.log(`Copying README.md to ${readmeDest}`);
fs.copyFileSync(readmeSrc, readmeDest);

console.log(`Copying LICENSE to ${licenseDest}`);
fs.copyFileSync(licenseSrc, licenseDest);

console.log(`Copying NOTICE to ${noticeDest}`);
fs.copyFileSync(noticeSrc, noticeDest);

console.log(`Copying THIRD_PARTY directory to ${thirdPartyDest}`);
if (fs.existsSync(thirdPartyDest)) {
  fs.rmSync(thirdPartyDest, { recursive: true, force: true });
}
fs.cpSync(thirdPartySrc, thirdPartyDest, { recursive: true });

console.log('Package preparation complete.');
