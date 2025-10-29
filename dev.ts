#!/usr/bin/env bun

import { spawn } from 'child_process';
import { existsSync } from 'fs';
import { join } from 'path';

const PROJECT_ROOT = process.cwd();
const BACKEND_DIR = join(PROJECT_ROOT, 'backend');
const FRONTEND_DIR = join(PROJECT_ROOT, 'frontend');

interface DependencyCheck {
  name: string;
  command: string;
  args: string[];
  versionFlag?: string;
}

const checks: DependencyCheck[] = [
  {
    name: 'Go',
    command: 'go',
    args: ['version'],
  },
  {
    name: 'Bun',
    command: 'bun',
    args: ['--version'],
  },
];

async function checkCommand(cmd: string, args: string[]): Promise<boolean> {
  return new Promise((resolve) => {
    const process = spawn(cmd, args, { stdio: 'pipe' });
    process.on('close', (code) => resolve(code === 0));
    process.on('error', () => resolve(false));
  });
}

async function checkDependencies(): Promise<boolean> {
  console.log('🔍 Checking dependencies...\n');
  
  let allGood = true;
  
  for (const check of checks) {
    const available = await checkCommand(check.command, check.args);
    if (available) {
      console.log(`✅ ${check.name} is available`);
    } else {
      console.log(`❌ ${check.name} is not available`);
      allGood = false;
    }
  }
  
  // Check Go modules
  if (existsSync(join(BACKEND_DIR, 'go.mod'))) {
    console.log('\n📦 Checking Go modules...');
    const goModAvailable = await checkCommand('go', ['mod', 'download'], { cwd: BACKEND_DIR });
    if (goModAvailable) {
      console.log('✅ Go modules are available');
    } else {
      console.log('❌ Failed to download Go modules');
      allGood = false;
    }
  }
  
  // Check Bun dependencies
  if (existsSync(join(FRONTEND_DIR, 'package.json'))) {
    console.log('\n📦 Checking frontend dependencies...');
    const nodeModulesExists = existsSync(join(FRONTEND_DIR, 'node_modules'));
    if (!nodeModulesExists) {
      console.log('📥 Installing frontend dependencies...');
      const installSuccess = await checkCommand('bun', ['install'], { cwd: FRONTEND_DIR });
      if (installSuccess) {
        console.log('✅ Frontend dependencies installed');
      } else {
        console.log('❌ Failed to install frontend dependencies');
        allGood = false;
      }
    } else {
      console.log('✅ Frontend dependencies are available');
    }
  }
  
  return allGood;
}

function runService(name: string, command: string, args: string[], cwd: string, color: string) {
  console.log(`\n🚀 Starting ${name}...`);
  
  const process = spawn(command, args, { 
    cwd, 
    stdio: 'inherit',
    env: { ...process.env }
  });
  
  process.stdout?.on('data', (data) => {
    console.log(`\x1b[${color}m[${name}]\x1b[0m ${data.toString().trim()}`);
  });
  
  process.stderr?.on('data', (data) => {
    console.error(`\x1b[${color}m[${name} ERROR]\x1b[0m ${data.toString().trim()}`);
  });
  
  process.on('close', (code) => {
    if (code !== 0) {
      console.error(`\x1b[${color}m[${name}]\x1b[0m Process exited with code ${code}`);
    }
  });
  
  process.on('error', (error) => {
    console.error(`\x1b[${color}m[${name} ERROR]\x1b[0m ${error.message}`);
  });
  
  return process;
}

async function main() {
  console.log('🛠️  Bronze Development Server\n');
  
  const depsOk = await checkDependencies();
  
  if (!depsOk) {
    console.log('\n❌ Dependency check failed. Please install missing dependencies.');
    process.exit(1);
  }
  
  console.log('\n✅ All dependencies satisfied!');
  console.log('🌟 Starting development servers...\n');
  
  // Start backend
  const backend = runService(
    'Backend',
    'go',
    ['run', 'main.go'],
    BACKEND_DIR,
    '32' // Green
  );
  
  // Wait a moment for backend to start
  await new Promise(resolve => setTimeout(resolve, 2000));
  
  // Start frontend
  const frontend = runService(
    'Frontend',
    'bun',
    ['run', 'dev'],
    FRONTEND_DIR,
    '34' // Blue
  );
  
  // Handle shutdown
  process.on('SIGINT', () => {
    console.log('\n\n🛑 Shutting down servers...');
    backend.kill('SIGINT');
    frontend.kill('SIGINT');
    process.exit(0);
  });
  
  console.log('\n🎉 Development servers are running!');
  console.log('📊 Backend: http://localhost:8080');
  console.log('🎨 Frontend: http://localhost:5173');
  console.log('\nPress Ctrl+C to stop all servers');
}

main().catch((error) => {
  console.error('❌ Failed to start development servers:', error);
  process.exit(1);
});