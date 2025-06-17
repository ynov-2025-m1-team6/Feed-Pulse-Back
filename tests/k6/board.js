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
    // Board metrics test
    board_test: {
      executor: 'constant-vus',
      vus: 15,
      duration: '30s',
      exec: 'boardTest',
    },
    // Progressive load test
    ramping_board: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 20 },
        { duration: '30s', target: 20 },
        { duration: '30s', target: 0 },
      ],
      exec: 'boardTest',
    },
    // High load test for dashboard
    dashboard_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 30 },
        { duration: '2m', target: 30 },
        { duration: '1m', target: 0 },
      ],
      exec: 'boardTest',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<1500'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
  },
};

// Login to get authentication token
function getAuthToken() {
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
    if (response.status === 200 && (response.headers['authorization'] || response.headers['Authorization'])) {
    return response.headers['authorization'] || response.headers['Authorization'];
  }
  
  return null;
}

// Test board metrics endpoint
function testBoardMetrics(token) {
  if (!token) return false;

  const url = `${baseUrl}/api/board/metrics`;

  const params = {
    headers: {
      'Authorization': token,
      'Content-Type': 'application/json',
    },
  };

  const response = http.get(url, params);

  return check(response, {
    'metrics status is 200': (r) => r.status === 200,
    'metrics response time < 1000ms': (r) => r.timings.duration < 1000,
    'metrics response has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data !== undefined;
      } catch (e) {
        return false;
      }
    },
    'metrics contains expected fields': (r) => {
      try {
        const body = JSON.parse(r.body);
        // Check if response contains typical dashboard metrics
        return body.data && (
          body.data.totalFeedbacks !== undefined ||
          body.data.averageRating !== undefined ||
          body.data.sentimentDistribution !== undefined ||
          Object.keys(body.data).length > 0
        );
      } catch (e) {
        return false;
      }
    },
  });
}

export function boardTest() {
  // Step 1: Login to get token
  const token = getAuthToken();
  
  if (!token) {
    errorRate.add(true);
    sleep(1);
    return;
  }

  sleep(0.5);

  // Step 2: Test board metrics
  const metricsSuccess = testBoardMetrics(token);

  // Error tracking
  errorRate.add(!metricsSuccess);

  sleep(1);
}

export function handleSummary(data) {
  return {
    [`./results/board-test-results-${environment}.json`]: JSON.stringify(data),
  };
}
