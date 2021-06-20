/*
This program helps reverse-engineering notionapi.

You give it the id of the Notion page and it'll download it
while recording requests and responses.

Summary of all requests is printed to stdout.

Api calls (/api/v3/) are logged to notion_api_trace.txt file
(pretty-printed body of POST data and pretty-printed JSON responses).

You need node.js. One time setup:
- cd tracenotion
- yarn (or: npm install)

To run manually:
- node ./tracenotion/trace.js <NOTION_PAGE_URL>

Or you can do:
- ./do/do.sh -trace <NOTION_PAGE_URL>

To access your private pages, set NOTION_TOKEN to the value
of token_v2 cookie on www.notion.so domain.
*/

const fs = require("fs");
const puppeteer = require("puppeteer");

const traceFilePath = "notion_api_trace.txt";

function trimStr(s, n) {
  if (s.length > n) {
    return s.substring(0, n) + "...";
  }
  return s;
}

function isApiRequest(url) {
  return url.includes("/api/v3/");
}

function ppjson(s) {
  try {
    js = JSON.parse(s);
    s = JSON.stringify(js, null, 2);
    return s;
  } catch {
    return s;
  }
}

let apiLog = [];

function logApiRR(method, url, status, reqBody, rspBody) {
  let s = `${method} ${status} ${url}`;
  if (!isApiRequest(url)) {
    apiLog.push(s);
    return;
  }
  apiLog.push(s);
  s = ppjson(reqBody);
  apiLog.push(s);
  s = ppjson(rspBody);
  apiLog.push(s);
  apiLog.push("-------------------------------");
}

function saveApiLog() {
  const s = apiLog.join("\n");
  fs.writeFileSync(traceFilePath, s);
  console.log(`Wrote api trace to ${traceFilePath}`);
}

let waitTime = 5 * 1000;
async function traceNotion(url) {
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  const token = process.env.NOTION_TOKEN || "";
  if (token !== "") {
    console.log("NOTION_TOKEN set, can access private pages");
    const c = {
      domain: "www.notion.so",
      name: "token_v2",
      value: token
    };
    await page.setCookie(c);
  } else {
    console.log("only public pages, NOTION_TOKEN env var not set");
  }
  await page.setRequestInterception(true);

  // those we don't want to log because they are not important
  function skipLogging(url) {
    const silenced = [
      "/api/v3/ping",
      "/appcache.html",
      "/loading-spinner.svg",
      "/api/v3/getUserAnalyticsSettings",
      "//analytics.pgncs.notion.so/analytics.js",
      "//api.pgncs.notion.so/",
      "//msgstore.www.notion.so/",
      "//www.notion.so/inter-ui-",
      "//www.notion.so/print.",
      "//www.notion.so/app-",
      "//www.notion.so/vendors~main-",
      "//www.notion.so/postRender-",
    ];
    for (let s of silenced) {
      if (url.includes(s)) {
        return true;
      }
    }
    return false;
  }

  function isBlacklisted(url) {
    const blacklisted = [
      "//amplitude.com/",
      "//fullstory.com/",
      ".intercom.io/",
      "//segment.io/",
      "//segment.com/",
      ".loggly.com/",
      "//js.intercomcdn.com",
      //"//analytics.pgncs.notion.so/analytics.js",
    ];
    for (let s of blacklisted) {
      if (url.includes(s)) {
        return true;
      }
    }
    return false;
  }

  page.on("request", request => {
    const url = request.url();
    if (isBlacklisted(url)) {
      request.abort();
      return;
    }
    request.continue();
  });

  page.on("requestfailed", request => {
    const url = request.url();
    if (isBlacklisted(url)) {
      // it was us who failed this request
      return;
    }
    console.log("request failed url:", url);
  });

  async function onResponse(response) {
    const request = response.request();
    let url = request.url();
    if (skipLogging(url)) {
      return;
    }
    let method = request.method();
    const postData = request.postData();

    // some urls are data urls and very long
    if (url.includes("data:")) {
      url = trimStr(url, 72);
    } else if (url.includes("msgstore.www.notion.so/")) {
      url = trimStr(url, 72);
    } else {
      // don't trim other urls, especially notion.so/image
    }
    const status = response.status();
    try {
      const d = await response.text();
      const dataLen = d.length;
      if (method === "GET") {
        // make the length same as POST
        method = "GET ";
      }
      console.log(`${method} ${status} ${url} size: ${dataLen}`);
      logApiRR(method, url, status, postData, d);
    } catch (ex) {
      console.log(`${method} ${status} ${url} ex: ${ex} FAIL !!!`);
    }
  }

  page.on("response", onResponse);

  await page.goto(url, { waitUntil: "networkidle2" });
  await page.waitFor(waitTime);

  await browser.close();
}

// a sample private url: https://www.notion.so/Things-15c47fa60c274ca2820629fb32c2be97
// a sample public url: https://www.notion.so/Test-text-4c6a54c68b3e4ea2af9cfaabcc88d58d

// first arg is "node"
// second arg is name of this script
// third is the first user argument
if (process.argv.length != 3) {
  console.log("Cell me as:");
  console.log("node ./tracenotion/trace.js <PAGE_URL>");
  console.log("e.g.:");
  console.log(
    "node ./tracenotion/trace.js https://www.notion.so/Test-text-4c6a54c68b3e4ea2af9cfaabcc88d58d"
  );
} else {
  async function doit() {
    const url = process.argv[2];
    await traceNotion(url);
    saveApiLog();
  }
  doit();
}
