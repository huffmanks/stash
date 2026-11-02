const os = require("os");

// Checks your macOS version and applies the esbuild downgrade
// if it detects you are on Big Sur (macOS 11).
module.exports = {
  hooks: {
    readPackage(pkg) {
      const isBigSur = os.platform() === "darwin" && os.release().startsWith("20.");

      if (isBigSur) {
        if (pkg.dependencies && pkg.dependencies.esbuild) {
          pkg.dependencies.esbuild = "0.24.2";
        }
        if (pkg.devDependencies && pkg.devDependencies.esbuild) {
          pkg.devDependencies.esbuild = "0.24.2";
        }
      }

      return pkg;
    },
  },
};
