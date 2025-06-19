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
    // Feedback operations test
    feedback_test: {
      executor: 'constant-vus',
      vus: 10,
      duration: '45s',
      exec: 'feedbackTest',
    },
    // Progressive load test
    ramping_feedback: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 15 },
        { duration: '45s', target: 15 },
        { duration: '30s', target: 0 },
      ],
      exec: 'feedbackTest',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.1'],
    errors: ['rate<0.1'],
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

// Test feedback upload endpoint
function testFeedbackUpload(token) {
  if (!token) return false;

  const url = `${baseUrl}/api/feedbacks/upload`;

  // Create a mock CSV content
  const csvContent = `name,email,feedback,rating
John Doe,john@example.com,"Great service, very satisfied!",5
Jane Smith,jane@example.com,"Could be better, some issues",3
Bob Johnson,bob@example.com,"Excellent experience",5`;

  const formData = {
    file: http.file(csvContent, 'feedback.csv', 'text/csv'),
  };

  const params = {
    headers: {
      'Authorization': token,
    },
  };

  const response = http.post(url, formData, params);

  return check(response, {
    'upload status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'upload response time < 3000ms': (r) => r.timings.duration < 3000,
    'upload response has message': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.message !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
}

// Test feedback fetch endpoint
function testFeedbackFetch(token) {
  if (!token) return false;

  const url = `${baseUrl}/api/feedbacks/fetch`;

  const payload = JSON.stringify({
    limit: 10,
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
    'fetch status is 200': (r) => r.status === 200,
    'fetch response time < 1000ms': (r) => r.timings.duration < 1000,
    'fetch response has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
}

// Test get feedbacks by user ID
function testGetFeedbacksByUserId(token) {
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
    'analyses status is 200': (r) => r.status === 200,
    'analyses response time < 1000ms': (r) => r.timings.duration < 1000,
    'analyses response has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
}

export function feedbackTest() {
  // Step 1: Login to get token
  const token = getAuthToken();

  if (!token) {
    errorRate.add(true);
    sleep(1);
    return;
  }

  sleep(0.5);

  // Step 2: Test feedback upload (every 3rd iteration to avoid too many uploads)
  let uploadSuccess = true;
  if (Math.random() < 0.33) {
    uploadSuccess = testFeedbackUpload(token);
    sleep(1);
  }

  // Step 3: Test feedback fetch
  const fetchSuccess = testFeedbackFetch(token);
  sleep(0.5);

  // Step 4: Test get feedbacks by user ID
  const analysesSuccess = testGetFeedbacksByUserId(token);

  // Overall success check
  const overallSuccess = uploadSuccess && fetchSuccess && analysesSuccess;
  errorRate.add(!overallSuccess);

  sleep(1);
}

export function handleSummary(data) {
  return {
    [`./results/feedback-test-results-${environment}.json`]: JSON.stringify(data),
  };
}
