import http from "k6/http";
import { check } from "k6";

export const options = {
  stages: [
    { duration: "10s", target: 50 },
    { duration: "30s", target: 50 },
    { duration: "10s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<50"],
  },
};

export default function () {
  const res = http.get("http://localhost:8080/bench", { redirects: 0 });
  check(res, { "status is 307": (r) => r.status === 307 });
}
