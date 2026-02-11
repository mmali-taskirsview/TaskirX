const tls = require("tls");

const hosts = [
  { name: "API Backend", host: "api.taskir.com" },
  { name: "Dashboard", host: "dashboard.taskir.com" },
  { name: "Bidding", host: "bidding.taskir.com" },
];

function probeHost({ name, host }) {
  return new Promise((resolve) => {
    const socket = tls.connect(
      {
        host,
        port: 443,
        servername: host,
        minVersion: "TLSv1.3",
        maxVersion: "TLSv1.3",
        rejectUnauthorized: false,
        timeout: 5000,
      },
      () => {
        const cert = socket.getPeerCertificate();
        const result = {
          name,
          host,
          success: true,
          protocol: socket.getProtocol(),
          subject: cert?.subject?.CN || "(unknown)",
          issuer: cert?.issuer?.CN || "(unknown)",
          validFrom: cert?.valid_from || "(unknown)",
          validTo: cert?.valid_to || "(unknown)",
        };
        socket.end();
        resolve(result);
      }
    );

    socket.on("error", (err) => {
      resolve({ name, host, success: false, error: err.message });
    });

    socket.on("timeout", () => {
      socket.destroy();
      resolve({ name, host, success: false, error: "Timeout" });
    });
  });
}

(async () => {
  console.log("TLS 1.3 probe for TaskirX endpoints");
  console.log("======================================");

  for (const entry of hosts) {
    const result = await probeHost(entry);
    if (result.success) {
      console.log(`${result.name}: OK (${result.protocol})`);
      console.log(`  Host     : ${result.host}`);
      console.log(`  Subject  : ${result.subject}`);
      console.log(`  Issuer   : ${result.issuer}`);
      console.log(`  Valid From: ${result.validFrom}`);
      console.log(`  Valid To  : ${result.validTo}`);
    } else {
      console.log(`${result.name}: FAIL`);
      console.log(`  Host : ${result.host}`);
      console.log(`  Error: ${result.error}`);
    }
    console.log("");
  }
})();
