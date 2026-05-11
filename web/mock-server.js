import http from 'http';

// Routes match the full path sent by the proxy (with /v1 prefix)
// The frontend calls /v1/... and vite proxy forwards to localhost:5174/v1/...
const responses = {
  // Auth
  'GET /v1/auth/setup-status': {
    code: 200, msg: 'ok', data: { setup_complete: true }
  },
  'POST /v1/auth/login': {
    code: 200, msg: 'ok', data: {
      token: 'mock-token-abc123',
      user: { id: '1', username: 'admin', role: 'admin', email: 'admin@example.com' }
    }
  },
  'GET /v1/auth/me': {
    code: 200, msg: 'ok', data: { id: '1', username: 'admin', role: 'admin', email: 'admin@example.com' }
  },

  // Dashboard
  'GET /v1/dashboard/stats': {
    code: 200, msg: 'ok', data: {
      containers: { total: 12, running: 8, stopped: 4 },
      machines: { total: 5 },
      networks: { total: 3, active: 3 },
      volumes: { total: 7, size: '2.4 GB' }
    }
  },
  'GET /v1/dashboard/resources/history': {
    code: 200, msg: 'ok', data: generateHistory()
  },

  // Remote machines
  'GET /v1/machines': {
    code: 200, msg: 'ok', data: {
      machines: [
        { id: 'm1', name: 'Production Server', status: 'online', host: '192.168.1.100' },
        { id: 'm2', name: 'Staging Server', status: 'online', host: '192.168.1.101' },
        { id: 'm3', name: 'Dev Server', status: 'offline', host: '192.168.1.102' },
        { id: 'm4', name: 'Backup Server', status: 'online', host: '192.168.1.103' }
      ]
    }
  },

  // Images
  'GET /v1/machines/local/images': {
    code: 200, msg: 'ok', data: {
      images: [
        { id: 'img1', repo_tags: ['nginx:latest'], size: 142000000 },
        { id: 'img2', repo_tags: ['node:20-alpine'], size: 125000000 }
      ]
    }
  },
  'GET /v1/machines/m1/images': {
    code: 200, msg: 'ok', data: {
      images: [
        { id: 'img3', repo_tags: ['ubuntu:22.04'], size: 77000000 }
      ]
    }
  },
  'GET /v1/machines/m2/images': {
    code: 200, msg: 'ok', data: {
      images: []
    }
  },
  'GET /v1/machines/m3/images': {
    code: 200, msg: 'ok', data: {
      images: []
    }
  },
  'GET /v1/machines/m4/images': {
    code: 200, msg: 'ok', data: {
      images: [
        { id: 'img4', repo_tags: ['busybox:latest'], size: 4200000 }
      ]
    }
  },

  // Compose projects
  'GET /v1/machines/m1/compose': {
    code: 200, msg: 'ok', data: {
      projects: [
        { id: 'c1', name: 'web-stack', status: 'running' },
        { id: 'c2', name: 'monitoring', status: 'running' }
      ]
    }
  },
  'GET /v1/machines/m2/compose': {
    code: 200, msg: 'ok', data: {
      projects: [
        { id: 'c3', name: 'staging-app', status: 'running' }
      ]
    }
  },
  'GET /v1/machines/m3/compose': {
    code: 200, msg: 'ok', data: {
      projects: []
    }
  },
  'GET /v1/machines/m4/compose': {
    code: 200, msg: 'ok', data: {
      projects: [
        { id: 'c4', name: 'backup-service', status: 'stopped' }
      ]
    }
  },
};

function generateHistory() {
  const points = [];
  const now = Date.now();
  for (let i = 60; i >= 0; i--) {
    const t = new Date(now - i * 2000);
    points.push({
      time: t.toLocaleTimeString('en-US', { hour12: false }),
      cpu: Math.round(Math.random() * 60 + 20),
      memory: Math.round(Math.random() * 30 + 40),
      disk: Math.round(Math.random() * 10 + 45),
    });
  }
  return points;
}

const server = http.createServer((req, res) => {
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');

  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  const url = new URL(req.url, `http://${req.headers.host}`);
  const path = url.pathname;
  const method = req.method;
  const key = `${method} ${path}`;

  let response = responses[key];

  // Fallback: try partial match for paths with query params
  if (!response) {
    const basePath = path.includes('?') ? path.split('?')[0] : path;
    for (const [k, v] of Object.entries(responses)) {
      const [m, p] = k.split(' ');
      if (method === m) {
        // Match by prefix (for history with query params like ?minutes=5)
        if (basePath.startsWith(p) || p.startsWith(basePath)) {
          response = v;
          break;
        }
      }
    }
  }

  if (response) {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(response));
  } else {
    console.log(`[404] ${method} ${path}`);
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ code: 200, msg: 'ok', data: null }));
  }
});

const PORT = 5174;
server.listen(PORT, () => {
  console.log(`Mock server running on http://localhost:${PORT}`);
  Object.keys(responses).forEach(k => console.log(`  ${k}`));
});
