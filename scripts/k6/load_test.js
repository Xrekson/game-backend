import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep, group } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// ==============================================================================
// Grafana k6 Distributed Game Backend Stress Test
// Hardware Profile: 2 vCPU, 4GB RAM Simulation (via Linux cgroups v2)
// SLO Targets:
//   - p(95) < 750ms (0.75s)
//   - p(99) < 2000ms (2.0s)
//   - Failure rate < 2%
// ==============================================================================

const GATEWAY_URL = __ENV.GATEWAY_URL || 'http://localhost:8080';
const WS_URL = __ENV.WS_URL || 'ws://localhost:8081/ws';
const P95_TARGET = __ENV.P95_MS || 750;
const P99_TARGET = __ENV.P99_MS || 2000;

// Custom metrics
const wsMessageSuccess = new Counter('ws_messages_sent');
const wsConnectionSuccess = new Rate('ws_connect_success');
const catalogDuration = new Trend('catalog_read_duration', true);
const claimDuration = new Trend('claim_income_duration', true);
const authDuration = new Trend('auth_register_duration', true);

export const options = {
    scenarios: {
        api_stress: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '10s', target: 20 },  // Ramp-up to 20 users
                { duration: '30s', target: 50 },  // Peak load: 50 concurrent users
                { duration: '10s', target: 75 },  // Stress spike to 75 users
                { duration: '10s', target: 0 },   // Graceful ramp-down
            ],
            gracefulRampDown: '5s',
        },
    },
    thresholds: {
        // Global HTTP latency thresholds requested: p95 < 0.75s, p99 < 2.0s
        'http_req_duration': [
            `p(95)<${P95_TARGET}`,
            `p(99)<${P99_TARGET}`,
        ],
        // Error rate must remain under 2%
        'http_req_failed': ['rate<0.02'],

        // Component-specific SLO thresholds
        'catalog_read_duration': [`p(95)<${P95_TARGET}`],
        'claim_income_duration': [`p(95)<${P95_TARGET}`],
        'http_req_duration{endpoint:health}': ['p(95)<100'],
    },
};

// Setup: Run once before VUs start. Registers a shared authenticated test account.
export function setup() {
    console.log(`Checking API Gateway health at ${GATEWAY_URL}/health...`);
    const healthRes = http.get(`${GATEWAY_URL}/health`);
    if (healthRes.status !== 200) {
        throw new Error(`Gateway unreachable (HTTP ${healthRes.status}): ${healthRes.body}`);
    }

    const uniqueId = Math.random().toString(36).substring(2, 10);
    const username = `k6_lead_${uniqueId}`;
    const email = `${username}@k6benchmark.local`;
    const password = 'Password123!';

    const regPayload = JSON.stringify({ username, email, password });
    const regRes = http.post(`${GATEWAY_URL}/api/v1/auth/register`, regPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    let token = '';
    let playerId = '';

    if (regRes.status === 200) {
        const body = JSON.parse(regRes.body);
        token = body.token || '';
        playerId = body.user?.id || '';
    } else {
        // Fallback to login
        const loginRes = http.post(`${GATEWAY_URL}/api/v1/auth/login`, JSON.stringify({ username, password }), {
            headers: { 'Content-Type': 'application/json' },
        });
        if (loginRes.status === 200) {
            const body = JSON.parse(loginRes.body);
            token = body.token || '';
            playerId = body.user?.id || '';
        }
    }

    console.log(`Setup complete. Authenticated as ${username} (Player ID: ${playerId})`);
    return { token, playerId };
}

// VU Execution Loop
export default function (data) {
    const authHeaders = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.token}`,
    };

    // 1. Gateway Health Baseline
    group('01_Gateway_Health', () => {
        const res = http.get(`${GATEWAY_URL}/health`, {
            tags: { endpoint: 'health' },
        });
        check(res, {
            'health status is 200': (r) => r.status === 200,
        });
    });

    // 2. Cassandra Catalog Read (Gateway -> Inventory gRPC -> Cassandra)
    group('02_Cassandra_Catalog_Read', () => {
        const res = http.get(`${GATEWAY_URL}/api/v1/inventory/templates`, {
            headers: authHeaders,
            tags: { endpoint: 'templates' },
        });
        catalogDuration.add(res.timings.duration);
        check(res, {
            'catalog status is 200': (r) => r.status === 200,
        });
    });

    // 3. PostgreSQL Player Profile & Economy (Gateway -> gRPC -> Postgres ACID)
    group('03_Postgres_Economy_Claim', () => {
        // Fetch balance
        const balRes = http.get(`${GATEWAY_URL}/api/v1/economy/balance`, {
            headers: authHeaders,
            tags: { endpoint: 'balance' },
        });
        check(balRes, {
            'balance status is 200': (r) => r.status === 200,
        });

        // Trigger income claim transaction
        const claimRes = http.post(`${GATEWAY_URL}/api/v1/economy/claim`, '{}', {
            headers: authHeaders,
            tags: { endpoint: 'claim' },
        });
        claimDuration.add(claimRes.timings.duration);
        check(claimRes, {
            'claim status is 200': (r) => r.status === 200,
        });
    });

    // 4. Marketplace Listings (Gateway -> Trading gRPC -> Postgres)
    group('04_Marketplace_Listings', () => {
        const res = http.get(`${GATEWAY_URL}/api/v1/trading/listings?status=ACTIVE`, {
            headers: authHeaders,
            tags: { endpoint: 'trading' },
        });
        check(res, {
            'trading listings status is 200': (r) => r.status === 200,
        });
    });

    // 5. Auth / Bcrypt Stress (Simulate 5% registration traffic)
    if (Math.random() < 0.05) {
        group('05_Auth_Registration_Bcrypt', () => {
            const randUser = `user_${Math.random().toString(36).substring(2, 9)}`;
            const payload = JSON.stringify({
                username: randUser,
                email: `${randUser}@test.local`,
                password: 'Password123!',
            });
            const regRes = http.post(`${GATEWAY_URL}/api/v1/auth/register`, payload, {
                headers: { 'Content-Type': 'application/json' },
                tags: { endpoint: 'auth_register' },
            });
            authDuration.add(regRes.timings.duration);
            check(regRes, {
                'registration status is 200': (r) => r.status === 200,
            });
        });
    }

    // 6. WebSocket Movement Stream (World Service :8081)
    if (Math.random() < 0.20) {
        group('06_WebSocket_Movement', () => {
            const vuPlayerId = `k6_player_${__VU}_${Math.floor(Math.random() * 1000)}`;
            const wsTarget = `${WS_URL}?player_id=${vuPlayerId}`;

            const res = ws.connect(wsTarget, {}, function (socket) {
                wsConnectionSuccess.add(1);

                socket.on('open', () => {
                    // Send 3 rapid movement packets to test coordinate validation
                    for (let i = 0; i < 3; i++) {
                        const moveMsg = JSON.stringify({
                            type: 'move',
                            payload: {
                                map_id: 'safe_zone_1',
                                x: 100.0 + i * 5,
                                y: 150.0 + i * 2,
                            },
                        });
                        socket.send(moveMsg);
                        wsMessageSuccess.add(1);
                        sleep(0.05);
                    }
                    socket.close();
                });

                socket.on('error', (err) => {
                    wsConnectionSuccess.add(0);
                });
            });

            check(res, {
                'websocket connection established': (r) => r && r.status === 101,
            });
        });
    }

    sleep(0.2); // Pacing between iterations
}

// Custom Markdown & JSON Summary Generator
export function handleSummary(data) {
    const httpDuration = data.metrics['http_req_duration']?.values || {};
    const httpReqs = data.metrics['http_reqs']?.values || {};
    const httpFailed = data.metrics['http_req_failed']?.values || {};
    const catalog = data.metrics['catalog_read_duration']?.values || {};
    const claim = data.metrics['claim_income_duration']?.values || {};
    const auth = data.metrics['auth_register_duration']?.values || {};
    const wsSent = data.metrics['ws_messages_sent']?.values || {};

    const p95 = httpDuration['p(95)'] || 0;
    const p99 = httpDuration['p(99)'] || 0;
    const rps = httpReqs['rate'] || 0;
    const failRate = ((httpFailed['rate'] || 0) * 100).toFixed(2);

    const p95Pass = p95 <= P95_TARGET ? '✅ PASS' : '❌ FAIL';
    const p99Pass = p99 <= P99_TARGET ? '✅ PASS' : '❌ FAIL';

    const report = `# 📊 Grafana k6 Stress Test Report
**Hardware Profile:** 2 vCPU, 4GB RAM (Linux cgroups v2 constraint)  
**Executed At:** ${new Date().toISOString()}  
**Target SLOs:** p(95) < ${P95_TARGET}ms | p(99) < ${P99_TARGET}ms  

---

### 🎯 Key Performance Indicators (KPIs)

| Metric | Target SLO | Actual Result | Status |
|---|---|---|---|
| **Latency p(95)** | **< ${P95_TARGET} ms** | **${p95.toFixed(1)} ms** | ${p95Pass} |
| **Latency p(99)** | **< ${P99_TARGET} ms** | **${p99.toFixed(1)} ms** | ${p99Pass} |
| **Total Throughput** | Best Effort | **${rps.toFixed(1)} req/s** | ⚡ High Throughput |
| **Error Rate** | < 2.0% | **${failRate}%** | ${parseFloat(failRate) < 2 ? '✅ STABLE' : '⚠️ HIGH'} |
| **Total HTTP Requests** | - | **${httpReqs['count'] || 0}** requests | - |
| **WebSocket Messages** | - | **${wsSent['count'] || 0}** packets | 🎮 Real-Time 2D |

---

### 🔍 Architectural Path Latency Breakdown

| Architectural Path | Mean Latency | p(90) | p(95) | p(99) | Datastore / Service |
|---|---|---|---|---|---|
| **Global HTTP Average** | ${(httpDuration['avg'] || 0).toFixed(1)} ms | ${(httpDuration['p(90)'] || 0).toFixed(1)} ms | **${p95.toFixed(1)} ms** | **${p99.toFixed(1)} ms** | All Microservices |
| **Cassandra Catalog Reads** | ${(catalog['avg'] || 0).toFixed(1)} ms | ${(catalog['p(90)'] || 0).toFixed(1)} ms | ${(catalog['p(95)'] || 0).toFixed(1)} ms | ${(catalog['p(99)'] || 0).toFixed(1)} ms | Gateway → gRPC → Cassandra |
| **Postgres ACID Income Claims** | ${(claim['avg'] || 0).toFixed(1)} ms | ${(claim['p(90)'] || 0).toFixed(1)} ms | ${(claim['p(95)'] || 0).toFixed(1)} ms | ${(claim['p(99)'] || 0).toFixed(1)} ms | Gateway → gRPC → Postgres |
| **Bcrypt Registration (CPU Bound)**| ${(auth['avg'] || 0).toFixed(1)} ms | ${(auth['p(90)'] || 0).toFixed(1)} ms | ${(auth['p(95)'] || 0).toFixed(1)} ms | ${(auth['p(99)'] || 0).toFixed(1)} ms | Auth Service (2 Cores) |

---

> 💡 **LinkedIn Post Takeaway:**  
> *"Under an active load of 50–75 concurrent virtual users on a 2-core / 4GB RAM budget sandbox, the Go microservices maintained a p95 latency of ${p95.toFixed(1)}ms and p99 of ${p99.toFixed(1)}ms against our SLOs of ${P95_TARGET}ms and ${P99_TARGET}ms."*
`;

    return {
        'stdout': textSummary(data, { indent: ' ', enableColors: true }),
        'logs/k6-report.md': report,
        'logs/k6-summary.json': JSON.stringify(data, null, 2),
    };
}

// Simple fallback text summary for console
function textSummary(data, options) {
    const httpDuration = data.metrics['http_req_duration']?.values || {};
    const httpReqs = data.metrics['http_reqs']?.values || {};
    const httpFailed = data.metrics['http_req_failed']?.values || {};

    return `
================================================================================
📈 GRAFANA K6 BENCHMARK COMPLETED
================================================================================
Total Requests:   ${httpReqs['count'] || 0}
Throughput (RPS): ${(httpReqs['rate'] || 0).toFixed(2)} req/s
Failure Rate:     ${(((httpFailed['rate'] || 0) * 100)).toFixed(2)}%
Latency (Avg):    ${(httpDuration['avg'] || 0).toFixed(2)} ms
Latency p(95):    ${(httpDuration['p(95)'] || 0).toFixed(2)} ms (Target: < ${P95_TARGET} ms)
Latency p(99):    ${(httpDuration['p(99)'] || 0).toFixed(2)} ms (Target: < ${P99_TARGET} ms)
================================================================================
Full Markdown Report saved to: logs/k6-report.md
`;
}
