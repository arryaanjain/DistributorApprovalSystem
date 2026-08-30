module.exports = {
  apps: [
    {
      name: "distributor-ui",

      cwd: "/var/www/krescoindia.com/production/DistributorApprovalSystem/ui",

      script: "npm",
      args: "start -- -p 3000",

      env: {
        NODE_ENV: "production",
        PORT: 3000,
      },
    },

    {
      name: "distributor-api",

      cwd: "/var/www/krescoindia.com/production/DistributorApprovalSystem/api",

      script: "./distributor-api",

      env: {
        PORT: 8081,
      },
    },
  ],
};