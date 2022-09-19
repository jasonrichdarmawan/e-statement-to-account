// reference: https://unpkg.com/external-svg-loader@1.6.1/svg-loader.js

async function isCacheAvailable(key: string) {
  let value = await GetIndexedDB(key)

  if (!value) {
    return;
  }

  let item = JSON.parse(value)

  if (Date.now() > item.expiry) {
    DeleteIndexedDB(key);
    return;
  }

  return item.data
}

function setCache(src: string, data: string) {
  const dataToSet = JSON.stringify({
    data,
    expiry: Date.now() + 60 * 60 * 1000 * 24 * 30
  })
  SetIndexedDB(src, dataToSet)
}

function renderBody(elem: SVGSVGElement, body: string) {
  const parser = new DOMParser();
  const doc = parser.parseFromString(body, "text/html");
  const svg = doc.querySelector("svg");
  if (svg == null) {
    throw Error(`Document does not contain svg element`);
  }
  let script = svg.querySelector("script");
  if (script) {
    script.remove();
    console.warn(
      `found <script> from fetching ${elem.getAttribute("data-src")}`
    );
  }
  elem.removeAttribute("data-src");
  Array.from(svg.attributes).forEach((attribute) => {
    elem.setAttribute(attribute.name, attribute.value);
  });
  elem.innerHTML = svg.innerHTML;
}

var requestsInProgress: { [key: string]: boolean } = {};
var memoryCache: { [key: string]: string } = {};

async function renderIcon(elem: SVGSVGElement) {
  let src = elem.getAttribute("data-src");
  if (src == null) {
    throw Error("svg element does not contain data-src attribute");
  }

  const isCache = await isCacheAvailable(src);

  // Memory cache optimizes same icon requested multiple
  // times on the page
  if (memoryCache[src] || isCache) {
    const cache = memoryCache[src] || isCache;
    renderBody(elem, cache);
    return;
  }

  // If the same icon is being requested to rendered
  // avoid firig multiple XHRs
  if (requestsInProgress[src]) {
    setTimeout(() => renderIcon(elem), 20);
    return;
  }
  requestsInProgress[src] = true;

  fetch(src)
    .then((response) => {
      if (!response.ok) {
        throw Error(
          `Request for ${src} returned ${response.status} ${response.statusText}`
        );
      }
      return response.text();
    })
    .then((body) => {
      if (src == null) {
        throw Error("let src is null");
      }

      if (!body.startsWith("<svg")) {
        throw Error(`Resource ${src} returned an invaldi SVG file`);
      }

      setCache(src, body)

      memoryCache[src] = body;

      renderBody(elem, body);
    })
    .catch((e) => console.error(e))
    .finally(() => {
      if (src == null) {
        throw Error("let src is null");
      }
      delete requestsInProgress[src]
    });

}

document.querySelectorAll<SVGSVGElement>("svg[data-src]")
  .forEach(element => {
    renderIcon(element)
  })
