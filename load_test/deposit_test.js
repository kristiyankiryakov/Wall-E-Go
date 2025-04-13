import http from 'k6/http';
import { sleep, check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';


export let options = {
    stages: [
        { duration: '5s', target: 10 }, // Ramp up to 100 users
        { duration: '30s', target: 1000 }, // Stay at 10 users
        { duration: '10s', target: 0 },  // Ramp down
    ],
};

export default function () {
    const url = 'http://localhost:8080/transaction/deposit'; // Replace with your service URL
    const payload = JSON.stringify({
        amount: 100,
        wallet_id: "4afbf382-b0b9-41e8-a945-b07d620bd066",
        idempotency_key: uuidv4(),
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization' : "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDQ2Mjc5MDcsImlzcyI6ImF1dGgtc2VydmVyIiwic3ViIjowfQ.yg1uJ-QEKC5SsSP82ALqO2CxJFi_xim216FbE2MrY0c"
        },
    };

    let res = http.put(url, payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        // 'transaction processed': (r) => r.body.includes("success"), // adjust as needed
    });

    // Only log a few responses to avoid spam
    if (__VU === 1 && __ITER < 5) {
        console.log(`Response: ${res.body}`);
    }

    sleep(1); // simulate user wait time
}
