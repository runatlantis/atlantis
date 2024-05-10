import path from "path";
import fs from "fs";
import process from "process";

export const getVersion = () => {
  const versionFilepath = path.join(__dirname, "../../version.txt");
  try {
    const version = fs.readFileSync(versionFilepath, "utf8").trim();
    console.log(`Found version ${version} from ${versionFilepath}`);
    return version;
  } catch (error) {
    console.error(`Failed to find version from file ${versionFilepath}`, error);
    process.exit(1);
  }
};

const en = [
  { text: "Home", link: "/"},
  { text: "Guide", link: "/guide/introduction" },
  { text: "Docs", link: "/docs/installation-guide" },
  { text: "Contributing", link: "/contributing" },
  { text: "Blog", link: "https://medium.com/runatlantis" },
];

export { en };
