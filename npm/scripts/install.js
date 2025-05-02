#!/usr/bin/env node
const { Binary } = require("binary-install");
const path = require("path");

const getPlatform = () => {
  const platform = process.platform;
  if (platform === "win32") {
    return "windows";
  }
  if (platform === "darwin") {
    return "macOS";
  }
  return platform;
};

const getArch = () => {
  const arch = process.arch;
  if (arch === "x64") {
    return "x86_64";
  }
  if (arch === "arm64") {
    return "arm64";
  }
  return arch;
};

const install = () => {
  const platform = getPlatform();
  const arch = getArch();
  const version = require("../package.json").version;
  const name = "ncurl";
  
  // Ensure we're using the correct version format for the GitHub release
  // This handles both regular versions and pre-release versions
  const releaseVersion = version;
  
  const url = `https://github.com/stephenbyrne99/ncurl/releases/download/v${releaseVersion}/${name}_${releaseVersion}_${platform}_${arch}.tar.gz`;
  
  console.log(`Downloading ${name} version ${releaseVersion} for ${platform} ${arch}...`);
  
  const binary = new Binary(name, url, { installDirectory: path.join(__dirname, "../") });
  
  try {
    binary.install();
    console.log(`Successfully installed ncurl v${version} for ${platform} ${arch}`);
  } catch (error) {
    console.error(`Error installing ncurl: ${error.message}`);
    process.exit(1);
  }
};

install();
