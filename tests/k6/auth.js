import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// Environment configuration
const environments = {
  local: 'http://localhost:3000',
  staging: 'https://feed-pulse-api-dev.onrender.com',
  prod: 'https://feed-pulse-api.onrender.com'
};

// Get environment from environment variable, default to local
const environment = __ENV.ENVIRONMENT || 'local';
const baseUrl = environments[environment];

export const options = {
  scenarios: {
    // Complete auth flow test
    auth_flow_test: {
      executor: 'constant-vus',
      vus: 15,
      duration: '1m',
      exec: 'authFlowTest',
    },
    // Progressive load test
    ramping_auth: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 20 },
        { duration: '1m', target: 20 },
        { duration: '30s', target: 0 },
      ],
      exec: 'authFlowTest',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<1500'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
  },
};

// Test login functionality
function loginUser() {
  const url = `${baseUrl}/api/auth/login`;
  const payload = JSON.stringify({
    login: "ftecher3",
    password: "12345678"
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);
  
  const success = check(response, {
    'login status is 200': (r) => r.status === 200,
    'login response time < 1000ms': (r) => r.timings.duration < 1000,    'has authorization header': (r) => r.headers['authorization'] !== undefined || r.headers['Authorization'] !== undefined,
  });

  if (success && response.status === 200) {
    return response.headers['authorization'] || response.headers['Authorization'];
  }
  
  return null;
}

// Test user info endpoint
function getUserInfo(token) {
  if (!token) return false;

  const url = `${baseUrl}/api/auth/user`;
  const params = {
    headers: {
      'Authorization': token,
      'Content-Type': 'application/json',
    },
  };

  const response = http.get(url, params);

  return check(response, {
    'user info status is 200': (r) => r.status === 200,
    'user info response time < 500ms': (r) => r.timings.duration < 500,
    'response contains user data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
}

// Test logout functionality
function logoutUser(token) {
  if (!token) return false;

  const url = `${baseUrl}/api/auth/logout`;
  const params = {
    headers: {
      'Authorization': token,
      'Content-Type': 'application/json',
    },
  };

  const response = http.get(url, params);

  return check(response, {
    'logout status is 200': (r) => r.status === 200,
    'logout response time < 500ms': (r) => r.timings.duration < 500,
  });
}

export function authFlowTest() {
  // Step 1: Login
  const token = loginUser();
  
  if (token) {
    sleep(0.5);
    
    // Step 2: Get user info
    const userInfoSuccess = getUserInfo(token);
    
    sleep(0.5);
    
    // Step 3: Logout
    const logoutSuccess = logoutUser(token);
    
    // Overall flow check
    const flowSuccess = token && userInfoSuccess && logoutSuccess;
    errorRate.add(!flowSuccess);
  } else {
    errorRate.add(true);
  }

  sleep(1);
}

export function handleSummary(data) {
  return {
    [`./results/auth-test-results-${environment}.json`]: JSON.stringify(data),
  };
}
