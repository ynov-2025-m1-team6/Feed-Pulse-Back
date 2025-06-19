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
    // Test de charge de base
    constant_load: {
      executor: 'constant-vus',
      vus: 20,
      duration: '30s',
      exec: 'pingTest',
    },
    // Test avec montée en charge progressive
    ramping_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },
        { duration: '30s', target: 20 },
        { duration: '30s', target: 0 },
      ],
      exec: 'pingTest',
    },
    // Test de stress
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 30 },
        { duration: '2m', target: 30 },
        { duration: '1m', target: 0 },
      ],
      exec: 'pingTest',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],   // Moins de 1% d'erreurs
  },
};

export function pingTest() {
  const response = http.get(`${baseUrl}/ping`);

  const checkResult = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(!checkResult);

  sleep(1);
}

export function handleSummary(data) {
  return {
    [`./results/ping-test-results-${environment}.json`]: JSON.stringify(data),
  };
}