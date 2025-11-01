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

async function checkCommand(cmd: string, args: string[], options?: { cwd?: string }): Promise<boolean> {
  return new Promise((resolve) => {
    const process = spawn(cmd, args, { 
      stdio: 'pipe',
      cwd: options?.cwd 
    });
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
    try {
      // Just check if go.mod is readable and go command works
      const goModExists = existsSync(join(BACKEND_DIR, 'go.mod'));
      if (goModExists) {
        console.log('✅ Go modules configuration found');
        // Try to tidy modules to ensure they're ready
        await checkCommand('go', ['mod', 'tidy'], { cwd: BACKEND_DIR });
        console.log('✅ Go modules are ready');
      } else {
        console.log('❌ go.mod not found');
        allGood = false;
      }
    } catch (error) {
      console.log('❌ Error checking Go modules:', error);
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
  
  const childProcess = spawn(command, args, { 
    cwd, 
    stdio: 'inherit',
    env: { ...process.env }
  });
  
  childProcess.stdout?.on('data', (data) => {
    console.log(`\x1b[${color}m[${name}]\x1b[0m ${data.toString().trim()}`);
  });
  
  childProcess.stderr?.on('data', (data) => {
    console.error(`\x1b[${color}m[${name} ERROR]\x1b[0m ${data.toString().trim()}`);
  });
  
  childProcess.on('close', (code) => {
    if (code !== 0) {
      console.error(`\x1b[${color}m[${name}]\x1b[0m Process exited with code ${code}`);
    }
  });
  
  childProcess.on('error', (error) => {
    console.error(`\x1b[${color}m[${name} ERROR]\x1b[0m ${error.message}`);
  });
  
  return childProcess;
}

async function killPorts() {
  console.log('🔧 Checking for processes on ports 8060 and 8070...');
  
  try {
    // Kill processes on port 8060 (backend)
    await checkCommand('lsof', ['-ti', ':8060']).then(async (pidExists) => {
      if (pidExists) {
        console.log('🔧 Killing processes on port 8060...');
        await checkCommand('kill', ['-9', '$(lsof -ti :8060)']);
      }
    });
    
    // Kill processes on port 8070 (frontend)
    await checkCommand('lsof', ['-ti', ':8070']).then(async (pidExists) => {
      if (pidExists) {
        console.log('🔧 Killing processes on port 8070...');
        await checkCommand('kill', ['-9', '$(lsof -ti :8070)']);
      }
    });
    
    // Alternative approach using pkill if lsof doesn't work
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Try to kill any remaining Go processes on port 8060
    await checkCommand('pkill', ['-f', 'go run main.go']);
    
    // Try to kill any remaining Vite processes on port 8070
    await checkCommand('pkill', ['-f', 'vite']);
    
    console.log('✅ Port cleanup completed');
  } catch (error) {
    console.log('⚠️  Port cleanup encountered an error, but continuing...');
  }
}

async function main() {
  console.log('🛠️  Bronze Development Server\n');
  
  const depsOk = await checkDependencies();
  
  if (!depsOk) {
    console.log('\n❌ Dependency check failed. Please install missing dependencies.');
    process.exit(1);
  }
  
  console.log('\n✅ All dependencies satisfied!');
  
  // Kill existing processes on ports
  await killPorts();
  
  console.log('🌟 Starting development servers...\n');
  
  // Start backend with auto-restart on crash
  let backend: ReturnType<typeof spawn> | null = null;

  function startBackend() {
    backend = runService(
      'Backend',
      'go',
      ['run', 'main.go'],
      BACKEND_DIR,
      '32' // Green
    );

    backend.on('close', (code) => {
      if (code !== 0 && code !== null) {
        console.log('\n🔄 Backend crashed, restarting in 1 second...');
        setTimeout(startBackend, 1000);
      }
    });
  }

  startBackend();

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
  const originalProcess = process;
  originalProcess.on('SIGINT', () => {
    console.log('\n\n🛑 Shutting down servers...');
    if (backend) backend.kill('SIGINT');
    frontend.kill('SIGINT');
    originalProcess.exit(0);
  });
  
  console.log('\n🎉 Development servers are running!');
  console.log('📊 Backend: http://localhost:8060');
  console.log('🎨 Frontend: http://localhost:8070');
  console.log('\nPress Ctrl+C to stop all servers');
}

main().catch((error) => {
  console.error('❌ Failed to start development servers:', error);
  process.exit(1);
});