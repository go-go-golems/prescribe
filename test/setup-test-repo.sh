#!/bin/bash
set -e

# Setup test repository for pr-builder testing

TEST_DIR="/tmp/pr-builder-test-repo"

echo "Creating test repository at $TEST_DIR..."

# Clean up if exists
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Initialize git repo
git init -b master 2>/dev/null || {
  git init
  git checkout -b master
}
git config user.email "test@example.com"
git config user.name "Test User"

# Create main branch with initial files
echo "Creating main branch..."

# Create directory structure
mkdir -p src/auth tests

cat > README.md << 'EOF'
# Test Project

This is a test project for PR Builder.
EOF

cat > src/auth/login.ts << 'EOFTS'
export async function login(
  email: string,
  password: string
) {
  // Simple login function
  const user = await db.findUser(email);
  if (!user) {
    throw new Error('User not found');
  }
  
  const valid = await comparePassword(password, user.passwordHash);
  if (!valid) {
    throw new Error('Invalid password');
  }
  
  return createSession(user);
}
EOFTS

cat > src/auth/middleware.ts << 'EOFTS'
export function authMiddleware(req, res, next) {
  const token = req.headers.authorization;
  if (!token) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  
  next();
}
EOFTS

cat > tests/auth.test.ts << 'EOFTS'
describe('auth', () => {
  it('should login successfully', async () => {
    const result = await login('test@example.com', 'password');
    expect(result).toBeDefined();
  });
});
EOFTS

# Commit initial state
git add .
git commit -m "Initial commit"

# Create feature branch
echo "Creating feature/user-auth branch..."
git checkout -b feature/user-auth

# Make changes to files
cat > src/auth/login.ts << 'EOFTS'
export async function login(
  email: string,
  password: string
) {
  // Validate input
  if (!email || !password) {
    throw new Error('Missing credentials');
  }
  
  // Hash password before compare
  const hash = await bcrypt.hash(password, 10);
  
  // Enhanced login function with validation
  const user = await db.findUser(email);
  if (!user) {
    throw new Error('User not found');
  }
  
  const valid = await comparePassword(password, user.passwordHash);
  if (!valid) {
    throw new Error('Invalid password');
  }
  
  // Log successful login
  await auditLog.log('login', { userId: user.id, email: user.email });
  
  return createSession(user);
}
EOFTS

cat > src/auth/middleware.ts << 'EOFTS'
import { verifyToken } from './token';
import { AuditLog } from './audit';

export function authMiddleware(req, res, next) {
  const token = req.headers.authorization;
  if (!token) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  
  try {
    // Verify token
    const decoded = verifyToken(token);
    req.user = decoded;
    
    // Log access
    AuditLog.log('api_access', {
      userId: decoded.id,
      path: req.path,
      method: req.method,
      ip: req.ip
    });
    
    next();
  } catch (error) {
    return res.status(401).json({ error: 'Invalid token' });
  }
}

export function requireRole(role: string) {
  return (req, res, next) => {
    if (!req.user || req.user.role !== role) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    next();
  };
}
EOFTS

cat > tests/auth.test.ts << 'EOFTS'
import { login } from '../src/auth/login';
import { authMiddleware } from '../src/auth/middleware';

describe('auth', () => {
  describe('login', () => {
    it('should login successfully with valid credentials', async () => {
      const result = await login('test@example.com', 'password123');
      expect(result).toBeDefined();
      expect(result.token).toBeDefined();
    });
    
    it('should reject invalid credentials', async () => {
      await expect(login('test@example.com', 'wrong')).rejects.toThrow('Invalid password');
    });
    
    it('should validate input', async () => {
      await expect(login('', 'password')).rejects.toThrow('Missing credentials');
    });
  });
  
  describe('middleware', () => {
    it('should allow authenticated requests', async () => {
      const req = { headers: { authorization: 'valid-token' } };
      const res = {};
      const next = jest.fn();
      
      authMiddleware(req, res, next);
      expect(next).toHaveBeenCalled();
    });
  });
});
EOFTS

# Commit changes
git add .
git commit -m "feat: enhance authentication with validation and audit logging

- Add input validation to login function
- Implement token verification in middleware
- Add audit logging for security events
- Expand test coverage for auth flows"

echo ""
echo "Test repository created successfully at: $TEST_DIR"
echo ""
echo "Branches:"
git branch -a
echo ""
echo "Files changed between master and feature/user-auth:"
git diff --stat master...feature/user-auth
