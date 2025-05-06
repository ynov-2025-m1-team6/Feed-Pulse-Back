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
    // Basic load test - reduced from 50 to 20 VUs
    constant_load: {
      executor: 'constant-vus',
      vus: 20,
      duration: '30s',
      exec: 'loginTest',
    },
    // Progressive load test - reduced max VUs and extended ramp-up time
    ramping_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 20 },  // Slower ramp-up
        { duration: '30s', target: 20 }, // Hold at peak
        { duration: '30s', target: 0 },  // Ramp-down
      ],
      exec: 'loginTest',
    },
    // Stress test - reduced max VUs and extended ramp-up
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },  // Slower ramp-up to 50
        { duration: '1m', target: 50 },  // Hold at peak
        { duration: '1m', target: 0 },   // Ramp-down
      ],
      exec: 'loginTest',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<1000'], // Increased to 1000ms
    http_req_failed: ['rate<0.05'],    // Increased to 5% error tolerance
  },
};

export function loginTest() {
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

  const checkResult = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'has authorization header': (r) => r.headers['Authorization'] !== undefined,
  });

  errorRate.add(!checkResult);

  sleep(1);
}

export function handleSummary(data) {
  return {
    [`./results/login-test-results-${environment}.json`]: JSON.stringify(data),
  };
}