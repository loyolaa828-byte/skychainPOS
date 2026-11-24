const http = require("http");
const fs = require("fs");
const path = require("path");

const ethersPath = "/workspaces/skychainPOS/evm/node_modules/ethers/dist/ethers.umd.min.js";

const server = http.createServer((req, res) => {
  if (req.url === "/rpc" && req.method === "POST") {
    const proxy = http.request(
      { hostname: "127.0.0.1", port: 8545, path: "/", method: "POST", headers: { "Content-Type": "application/json" } },
      (pRes) => {
        res.writeHead(pRes.statusCode, { "Content-Type": "application/json" });
        pRes.pipe(res);
      }
    );
    proxy.on("error", () => {
      res.writeHead(502, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "EVM RPC offline" }));
    });
    req.pipe(proxy);
    return;
  }
  if (req.url === "/ethers.js") {
    res.writeHead(200, { "Content-Type": "application/javascript" });
    res.end(fs.readFileSync(ethersPath));
    return;
  }
  if (req.url === "/contracts.json") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(fs.readFileSync(path.join(__dirname, "contracts.json"), "utf8"));
    return;
  }
  res.writeHead(200, { "Content-Type": "text/html" });
  res.end(fs.readFileSync(path.join(__dirname, "index.html"), "utf8"));
});

server.listen(4000, "0.0.0.0", () => console.log("SkynetScan Explorer + RPC Proxy live on :4000"));
