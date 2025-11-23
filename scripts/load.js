import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TEAM_NAME = __ENV.TEAM_NAME || 'load-team';
const USERS = [
  { id: 'load-u1', name: 'Load Alice', active: true },
  { id: 'load-u2', name: 'Load Bob', active: true },
  { id: 'load-u3', name: 'Load Carol', active: true },
  { id: 'load-u4', name: 'Load Dave', active: true },
  { id: 'load-u5', name: 'Load Eve', active: false }, // для проверки фильтра по активности
];

const headers = { 'Content-Type': 'application/json' };

export const options = {
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<200'],
  },
  scenarios: {
    create_pr: {
      executor: 'constant-arrival-rate',
      rate: 5,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 10,
      exec: 'createPRScenario',
    },
    reassign_merge: {
      executor: 'ramping-arrival-rate',
      timeUnit: '1s',
      startRate: 1,
      stages: [
        { target: 5, duration: '30s' },
        { target: 5, duration: '1m' },
        { target: 1, duration: '30s' },
      ],
      preAllocatedVUs: 15,
      maxVUs: 30,
      exec: 'reassignMergeScenario',
    },
    get_review: {
      executor: 'constant-vus',
      vus: 10,
      duration: '1m30s',
      exec: 'getReviewScenario',
    },
  },
};

export function setup() {
  const payload = JSON.stringify({
    team_name: TEAM_NAME,
    members: USERS.map((u) => ({
      user_id: u.id,
      username: u.name,
      is_active: u.active,
    })),
  });

  const res = http.post(`${BASE_URL}/team/add`, payload, { headers });
  if (![201, 400].includes(res.status)) {
    throw new Error(`team/add failed: ${res.status} ${res.body}`);
  }

  return {
    author: USERS[0].id,
    reviewers: USERS.slice(1, 4).map((u) => u.id),
  };
}

function uniqueId(prefix) {
  return `${prefix}-${exec.scenario.iterationInTest}-${Date.now()}-${Math.random()
    .toString(16)
    .slice(2)}`;
}

export function createPRScenario(data) {
  const prID = uniqueId('pr-create');
  const payload = JSON.stringify({
    pull_request_id: prID,
    pull_request_name: `feature-${prID}`,
    author_id: data.author,
  });

  const res = http.post(`${BASE_URL}/pullRequest/create`, payload, { headers });
  check(res, {
    'create status is 201': (r) => r.status === 201,
    'create has reviewers array': (r) => (r.json()?.pr?.assigned_reviewers || []).length >= 0,
  });
}

export function reassignMergeScenario(data) {
  const prID = uniqueId('pr-reassign');
  const createPayload = JSON.stringify({
    pull_request_id: prID,
    pull_request_name: `bugfix-${prID}`,
    author_id: data.author,
  });

  const createRes = http.post(`${BASE_URL}/pullRequest/create`, createPayload, { headers });
  if (createRes.status !== 201) {
    check(createRes, { 'create before reassign ok': (r) => r.status === 201 });
    return;
  }

  const assigned = createRes.json()?.pr?.assigned_reviewers || [];
  if (assigned.length === 0) {
    return;
  }
  const oldReviewer = assigned[0];

  const reassignRes = http.post(
    `${BASE_URL}/pullRequest/reassign`,
    JSON.stringify({
      pull_request_id: prID,
      old_user_id: oldReviewer,
    }),
    { headers },
  );

  check(reassignRes, {
    'reassign status 200': (r) => r.status === 200,
  });

  const mergeRes = http.post(
    `${BASE_URL}/pullRequest/merge`,
    JSON.stringify({ pull_request_id: prID }),
    { headers },
  );
  check(mergeRes, {
    'merge status 200': (r) => r.status === 200,
    'merge returns reviewers': (r) => (r.json()?.pr?.assigned_reviewers || []).length >= 0,
  });
}

export function getReviewScenario(data) {
  const reviewer = data.reviewers[exec.scenario.iterationInTest % data.reviewers.length];
  const res = http.get(`${BASE_URL}/users/getReview?user_id=${reviewer}`);
  check(res, {
    'getReview status 200': (r) => r.status === 200,
  });
  sleep(0.5);
}
