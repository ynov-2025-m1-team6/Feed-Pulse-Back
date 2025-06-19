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
    // Stress test - High load
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },   // Ramp up to 50 users
        { duration: '5m', target: 100 },  // Scale to 100 users
        { duration: '2m', target: 150 },  // Spike to 150 users
        { duration: '3m', target: 150 },  // Hold peak load
        { duration: '2m', target: 50 },   // Scale down
        { duration: '2m', target: 0 },    // Ramp down
      ],
      exec: 'stressTest',
    },
    // Spike test - Sudden load increase
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 10,
      stages: [
        { duration: '1m', target: 10 },   // Normal load
        { duration: '30s', target: 200 }, // Sudden spike
        { duration: '1m', target: 200 },  // Hold spike
        { duration: '30s', target: 10 },  // Return to normal
        { duration: '1m', target: 10 },   // Hold normal
      ],
      exec: 'stressTest',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.15'], // Allow higher error rate during stress
    errors: ['rate<0.15'],
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

// Stress test with mixed endpoints
export function stressTest() {
  // Randomly choose which endpoint to stress test
  const testChoice = Math.random();

  if (testChoice < 0.2) {
    // 20% - Test ping endpoint (lightest)
    testPingEndpoint();
  } else if (testChoice < 0.4) {
    // 20% - Test login endpoint
    testLoginEndpoint();
  } else if (testChoice < 0.6) {
    // 20% - Test authenticated user info
    testUserInfoEndpoint();
  } else if (testChoice < 0.8) {
    // 20% - Test feedback fetch
    testFeedbackEndpoint();
  } else {
    // 20% - Test board metrics
    testBoardEndpoint();
  }

  // Variable sleep to simulate real user behavior
  sleep(Math.random() * 1.5 + 0.5);
}

function testPingEndpoint() {
  const url = `${baseUrl}/ping`;
  const response = http.get(url);

  const success = check(response, {
    'ping status is 200': (r) => r.status === 200,
    'ping response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
}

function testLoginEndpoint() {
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
    'login response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  errorRate.add(!success);
}

function testUserInfoEndpoint() {
  const token = getAuthToken();
  if (!token) {
    errorRate.add(true);
    return;
  }

  const url = `${baseUrl}/api/auth/user`;
  const params = {
    headers: {
      'Authorization': token,
      'Content-Type': 'application/json',
    },
  };

  const response = http.get(url, params);

  const success = check(response, {
    'user info status is 200': (r) => r.status === 200,
    'user info response time < 1500ms': (r) => r.timings.duration < 1500,
  });

  errorRate.add(!success);
}

function testFeedbackEndpoint() {
  const token = getAuthToken();
  if (!token) {
    errorRate.add(true);
    return;
  }

  const url = `${baseUrl}/api/feedbacks/fetch`;
  const payload = JSON.stringify({
    limit: 10,
    offset: Math.floor(Math.random() * 50) // Random offset for variety
  });

  const params = {
    headers: {
      'Authorization': token,
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);

  const success = check(response, {
    'feedback fetch status is 200': (r) => r.status === 200,
    'feedback fetch response time < 2500ms': (r) => r.timings.duration < 2500,
  });

  errorRate.add(!success);
}

function testBoardEndpoint() {
  const token = getAuthToken();
  if (!token) {
    errorRate.add(true);
    return;
  }

  const url = `${baseUrl}/api/board/metrics`;

  const params = {
    headers: {
      'Authorization': token,
      'Content-Type': 'application/json',
    },
  };

  const response = http.get(url, params);

  const success = check(response, {
    'board metrics status is 200': (r) => r.status === 200,
    'board metrics response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  errorRate.add(!success);
}

export function handleSummary(data) {
  return {
    [`./results/stress-test-results-${environment}.json`]: JSON.stringify(data),
  };
}
