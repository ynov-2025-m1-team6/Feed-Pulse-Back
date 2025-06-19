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
        { duration: '30s', target: 10 },
        { duration: '45s', target: 15 },
        { duration: '30s', target: 0 },
      ],
      exec: 'feedbackTest',
    },
  },
  thresholds: {
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
  const csvContent = `[
    {
      "id": "fb_001",
      "date": "2025-04-14T10:30:00Z",
      "channel": "twitter",
      "text": "Le support client a été très réactif et efficace."
    },
    {
      "id": "fb_002",
      "date": "2025-04-14T11:00:00Z",
      "channel": "facebook",
      "text": "Je trouve les tarifs un peu élevés pour les fonctionnalités proposées."
    }
    ]`;

  const formData = {
    file: http.file(csvContent, 'feedback.json', 'application/json'),
  };

  const params = {
    headers: {
      'Authorization': token,
    },
  };

  const response = http.post(url, formData, params);

  return check(response, {
    'upload status is 200 or 201': (r) => r.status === 200 || r.status === 201,
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
