/*
This program helps reverse-engineering notionapi.
You give it the id of the Notion page and it'll download it
while recording all requests and responses to stdout.
You can then inspect the API calls to write a wrapper around
them.

You need node.js.

To run manually:
- cd tracenotion
- yarn (or: npm install)
- node trace_notion_apis.js <NOTION_PAGE_URL>

You probably want to save output to a file with:
node trace_notion_apis.js <NOTION_PAGE_URL> >trace.txt

Or you can do:
- ./do/do.sh -trace <NOTION_PAGE_URL>

To access your private pages, set NOTION_TOKEN to the value
of token_v2 cookie on www.notion.so domain.
*/

/*
Actually implement:
- get url from cmd-line
- do.sh support
- option -only-api which only shows /api/v3/* requests
*/

const puppeteer = require("puppeteer");

function trimStr(s, n) {
  if (s.length > n) {
    return s.substring(0, n) + "...";
  }
  return s;
}

function isApiRequest(url) {
  return url.Contains("/api/v3/");
}

function ppjson(s) {
  // TODO: pretty-print json
  return s;
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
  function isSilenced(url) {
    const silenced = [
      "/api/v3/ping",
      "/appcache.html",
      "/loading-spinner.svg",
      "/api/v3/getUserAnalyticsSettings"
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
      "amplitude.com/",
      "fullstory.com/",
      "intercom.io/",
      "segment.io/",
      "loggly.com/"
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

  page.on("response", response => {
    const request = response.request();
    let url = request.url();
    if (isSilenced(url)) {
      return;
    }
    const method = request.method();
    const postData = request.postData();

    // some urls are data urls and very long
    url = trimStr(url, 72);
    const status = response.status();
    response
      .text()
      .then(d => {
        const dataLen = d.length;
        console.log(`${method} ${url} ${status} size: ${dataLen}`);
        if (postData) {
          console.log(postData);
        }
      })
      .catch(reason => {
        console.log(`Failed to get response: ${method} ${url} ${status}`);
      });
  });

  await page.goto(url, { waitUntil: "networkidle2" });
  await page.waitFor(waitTime);

  await browser.close();
}

const url =
  "https://www.notion.so/Log-short-term-todo-e4d392caeef64b9286070c2ee712f725";
// const url = "https://www.notion.so/Test-page-all-c969c9455d7c4dd79c7f860f3ace6429";

traceNotion(url);
