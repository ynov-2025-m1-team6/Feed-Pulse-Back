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
    // Basic registration test
    register_test: {
      executor: 'constant-vus',
      vus: 10,
      duration: '30s',
      exec: 'registerTest',
    },
    // Progressive load test
    ramping_register: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 15 },
        { duration: '30s', target: 15 },
        { duration: '30s', target: 0 },
      ],
      exec: 'registerTest',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.1'],
    errors: ['rate<0.1'],
  },
};

export function registerTest() {
  // Generate unique user for each test iteration
  const uniqueId = Math.floor(Math.random() * 1000000);
  const url = `${baseUrl}/api/auth/register`;

  const payload = JSON.stringify({
    username: `testuser${uniqueId}`,
    email: `test${uniqueId}@example.com`,
    password: "testpassword123",
    firstName: "Test",
    lastName: "User"
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);

  const checkResult = check(response, {
    'status is 201 or 400': (r) => r.status === 201 || r.status === 400, // 400 for existing user
    'response time < 1000ms': (r) => r.timings.duration < 1000,
    'response has message': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.message !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!checkResult);

  sleep(1);
}

export function handleSummary(data) {
  return {
    [`./results/register-test-results-${environment}.json`]: JSON.stringify(data),
  };
}
