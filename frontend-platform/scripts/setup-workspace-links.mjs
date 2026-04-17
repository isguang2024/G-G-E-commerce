import {
  existsSync,
  lstatSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  symlinkSync,
} from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const rootDir = path.dirname(fileURLToPath(import.meta.url));
const workspaceRoot = path.dirname(rootDir);
const nodeModulesDir = path.join(workspaceRoot, 'node_modules');
const scanRoots = ['apps', 'docs', 'internal', 'packages', 'playground', 'scripts'];
const skipDirs = new Set([
  '.git',
  '.turbo',
  '.vite',
  '.vscode',
  'coverage',
  'dist',
  'node_modules',
]);
const supportedScopes = new Set(['@vben', '@vben-core']);

function collectPackageJsonFiles(startDir, files = []) {
  if (!existsSync(startDir)) {
    return files;
  }

  for (const entry of readdirSync(startDir, { withFileTypes: true })) {
    if (skipDirs.has(entry.name)) {
      continue;
    }

    const fullPath = path.join(startDir, entry.name);
    if (entry.isDirectory()) {
      collectPackageJsonFiles(fullPath, files);
      continue;
    }

    if (entry.isFile() && entry.name === 'package.json') {
      files.push(fullPath);
    }
  }

  return files;
}

function ensureScopeLink(packageName, packageDir) {
  const [scope, name] = packageName.split('/');
  if (!supportedScopes.has(scope) || !name) {
    return false;
  }

  const scopeDir = path.join(nodeModulesDir, scope);
  const linkPath = path.join(scopeDir, name);

  mkdirSync(scopeDir, { recursive: true });

  if (existsSync(linkPath)) {
    return false;
  }

  symlinkSync(packageDir, linkPath, process.platform === 'win32' ? 'junction' : 'dir');
  return true;
}

const packageFiles = scanRoots.flatMap((dir) =>
  collectPackageJsonFiles(path.join(workspaceRoot, dir)),
);

let created = 0;
for (const packageFile of packageFiles) {
  const packageDir = path.dirname(packageFile);
  const packageJson = JSON.parse(readFileSync(packageFile, 'utf8'));
  const packageName = packageJson?.name;

  if (typeof packageName !== 'string') {
    continue;
  }

  if (ensureScopeLink(packageName, packageDir)) {
    created += 1;
  }
}

const tsconfigLink = path.join(nodeModulesDir, '@vben', 'tsconfig');
if (!existsSync(tsconfigLink)) {
  throw new Error('workspace links setup failed: missing node_modules/@vben/tsconfig');
}

const linkStat = lstatSync(tsconfigLink);
if (!linkStat.isDirectory() && !linkStat.isSymbolicLink()) {
  throw new Error('workspace links setup failed: node_modules/@vben/tsconfig is not a directory link');
}

console.log(`[setup-workspace-links] ready, created ${created} links`);
