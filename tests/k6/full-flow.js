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
    // Complete application flow test
    full_flow_test: {
      executor: 'constant-vus',
      vus: 8,
      duration: '2m',
      exec: 'fullFlowTest',
    },
    // User journey simulation
    user_journey: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 12 },
        { duration: '3m', target: 12 },
        { duration: '1m', target: 0 },
      ],
      exec: 'fullFlowTest',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<2500'],
    http_req_failed: ['rate<0.08'],
    errors: ['rate<0.08'],
  },
};

// Test ping endpoint
function testPing() {
  const url = `${baseUrl}/ping`;
  const response = http.get(url);

  return check(response, {
    'ping status is 200': (r) => r.status === 200,
    'ping response time < 200ms': (r) => r.timings.duration < 200,
  });
}

// Login to get authentication token
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
  });
}

// Test feedback fetch
function fetchFeedbacks(token) {
  if (!token) return false;

  const url = `${baseUrl}/api/feedbacks/fetch`;

  const payload = JSON.stringify({
    limit: 5,
    offset: 0
  });

  const params = {
    headers: {
      'Authorization': token,
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);

  return check(response, {
    'fetch feedbacks status is 200': (r) => r.status === 200,
    'fetch feedbacks response time < 1000ms': (r) => r.timings.duration < 1000,
  });
}

// Test board metrics
function getBoardMetrics(token) {
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
    'board metrics status is 200': (r) => r.status === 200,
    'board metrics response time < 1000ms': (r) => r.timings.duration < 1000,
  });
}

// Test feedback analyses
function getFeedbackAnalyses(token) {
  if (!token) return false;

  const url = `${baseUrl}/api/feedbacks/analyses`;

  const params = {
    headers: {
      'Authorization': token,
      'Content-Type': 'application/json',
    },
  };

  const response = http.get(url, params);

  return check(response, {
    'feedback analyses status is 200': (r) => r.status === 200,
    'feedback analyses response time < 1000ms': (r) => r.timings.duration < 1000,
  });
}

// Test logout
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

export function fullFlowTest() {
  // Step 1: Test ping (health check)
  const pingSuccess = testPing();
  sleep(0.3);

  // Step 2: Login
  const token = loginUser();

  if (!token) {
    errorRate.add(true);
    sleep(2);
    return;
  }

  sleep(0.5);

  // Step 3: Get user info
  const userInfoSuccess = getUserInfo(token);
  sleep(0.5);

  // Step 4: Fetch feedbacks
  const fetchSuccess = fetchFeedbacks(token);
  sleep(0.7);

  // Step 5: Get board metrics (dashboard view)
  const metricsSuccess = getBoardMetrics(token);
  sleep(0.5);

  // Step 6: Get feedback analyses
  const analysesSuccess = getFeedbackAnalyses(token);
  sleep(0.5);

  // Step 7: Logout
  const logoutSuccess = logoutUser(token);

  // Overall flow success check
  const overallSuccess = pingSuccess && token && userInfoSuccess &&
                        fetchSuccess && metricsSuccess && analysesSuccess && logoutSuccess;

  errorRate.add(!overallSuccess);

  // Simulate user think time
  sleep(Math.random() * 2 + 1);
}

export function handleSummary(data) {
  return {
    [`./results/full-flow-test-results-${environment}.json`]: JSON.stringify(data),
  };
}
