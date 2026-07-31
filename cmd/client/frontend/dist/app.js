const deviceListView = document.querySelector("#device-list-view");
const deviceControlView = document.querySelector("#device-control-view");
const devices = document.querySelector("#devices");
const status = document.querySelector("#status");
const refresh = document.querySelector("#refresh");
const back = document.querySelector("#back");
const selectedDeviceName = document.querySelector("#selected-device-name");
const selectedDeviceDetails = document.querySelector("#selected-device-details");
const controlStatus = document.querySelector("#control-status");
const appsElement = document.querySelector("#apps");
const appCount = document.querySelector("#app-count");
const appControls = document.querySelector("#app-controls");

let renderedDevices = [];
let renderedApps = [];
let selectedAppID = null;

function escapeHTML(value) {
  const node = document.createElement("span");
  node.textContent = value || "";
  return node.innerHTML;
}

function renderDevices(found) {
  renderedDevices = found;

  if (!found.length) {
    devices.innerHTML = `
      <div class="empty">
        <strong>No Cast devices found</strong>
        <span>Check that this computer and your Chromecast are on the same network.</span>
      </div>`;
    return;
  }

  devices.innerHTML = found.map((device, index) => `
    <button class="device" type="button" data-device-index="${index}">
      <span class="device-icon" aria-hidden="true">◧</span>
      <span class="device-copy">
        <strong>${escapeHTML(device.name)}</strong>
        <span>${escapeHTML(device.model || "Google Cast device")}</span>
        <code>${escapeHTML(device.host)}:${device.port}</code>
      </span>
      <span class="state">${escapeHTML(device.connectionState)}</span>
    </button>`).join("");
}

function renderApps(found) {
  renderedApps = found || [];
  appCount.textContent = `${renderedApps.length} available`;

  if (!renderedApps.length) {
    appsElement.innerHTML = `
      <div class="empty compact">
        <strong>No available apps</strong>
        <span>This receiver did not report support for a known app.</span>
      </div>`;
    selectedAppID = null;
    renderAppControls();
    return;
  }

  if (!renderedApps.some((app) => app.id === selectedAppID)) {
    selectedAppID = null;
  }

  appsElement.innerHTML = renderedApps.map((app) => `
    <button class="app${app.id === selectedAppID ? " selected" : ""}" type="button" data-app-id="${escapeHTML(app.id)}">
      <span class="app-name">${app.running ? "⭐ " : ""}${escapeHTML(app.name)}</span>
      <span class="app-id">${escapeHTML(app.id)}</span>
    </button>`).join("");
  renderAppControls();
}

function renderAppControls() {
  const app = renderedApps.find((candidate) => candidate.id === selectedAppID);
  if (!app) {
    appControls.innerHTML = `
      <div class="empty app-empty">
        <strong>Select an app</strong>
        <span>Choose an application to see its controls.</span>
      </div>`;
    return;
  }

  const isYouTube = app.id === "233637DE" || app.id === "2C6A6E3D";
  const youtubeControls = isYouTube ? `
    <form id="youtube-form" class="youtube-controls">
      <label for="youtube-url">YouTube video URL</label>
      <div class="input-action">
        <input id="youtube-url" name="youtube-url" type="url" required placeholder="https://www.youtube.com/watch?v=…">
        <button type="submit">${app.running ? "Play video" : "Launch and play"}</button>
      </div>
    </form>` : "";
  const appAction = app.running ? `
    <button id="app-action" class="danger" type="button" data-action="terminate">Terminate app</button>` : (!isYouTube ? `
    <button id="app-action" type="button" data-action="launch">Launch app</button>` : "");

  appControls.innerHTML = `
    <div class="control-copy">
      <p class="eyebrow">${app.running ? "RUNNING" : "AVAILABLE"}</p>
      <h2>${escapeHTML(app.name)}</h2>
      <code>${escapeHTML(app.id)}</code>
      <p>${escapeHTML(app.statusText || (app.running ? "This app is currently active." : "This app can be launched on the selected device."))}</p>
    </div>
    <div class="control-actions">
      ${youtubeControls}
      ${appAction}
    </div>`;
}

function showDeviceList() {
  deviceControlView.hidden = true;
  deviceListView.hidden = false;
}

async function openDevice(device) {
  selectedAppID = null;
  selectedDeviceName.textContent = device.name;
  selectedDeviceDetails.textContent = `${device.model || "Google Cast device"} · ${device.host}:${device.port}`;
  deviceListView.hidden = true;
  deviceControlView.hidden = false;
  controlStatus.className = "status scanning";
  controlStatus.textContent = "Connecting and querying available apps…";
  renderedApps = [];
  appsElement.innerHTML = `
    <div class="empty compact">
      <strong>Querying apps…</strong>
      <span>Checking which known applications this receiver supports.</span>
    </div>`;
  appCount.textContent = "";
  renderAppControls();

  try {
    const found = await window.go.main.App.SelectDevice(device);
    renderApps(found);
    controlStatus.className = "status";
    controlStatus.textContent = `Connected to ${device.name}`;
  } catch (error) {
    controlStatus.className = "status error";
    controlStatus.textContent = `Unable to load device: ${error}`;
  }
}

async function runAppAction(button) {
  const app = renderedApps.find((candidate) => candidate.id === selectedAppID);
  if (!app) {
    return;
  }

  button.disabled = true;
  controlStatus.className = "status scanning";
  controlStatus.textContent = `${app.running ? "Terminating" : "Launching"} ${app.name}…`;
  try {
    const found = app.running
      ? await window.go.main.App.TerminateApp(app.id)
      : await window.go.main.App.LaunchApp(app.id);
    renderApps(found);
    controlStatus.className = "status";
    controlStatus.textContent = `${app.name} ${app.running ? "terminated" : "launched"}`;
  } catch (error) {
    controlStatus.className = "status error";
    controlStatus.textContent = `App action failed: ${error}`;
    button.disabled = false;
  }
}

async function playYouTube(form) {
  const url = new FormData(form).get("youtube-url");
  const button = form.querySelector("button[type=submit]");
  button.disabled = true;
  controlStatus.className = "status scanning";
  controlStatus.textContent = "Sending video to YouTube…";
  try {
    const found = await window.go.main.App.PlayYouTube(selectedAppID, url);
    renderApps(found);
    controlStatus.className = "status";
    controlStatus.textContent = "Video sent to YouTube";
  } catch (error) {
    controlStatus.className = "status error";
    controlStatus.textContent = `YouTube playback failed: ${error}`;
    button.disabled = false;
  }
}

async function scan() {
  refresh.disabled = true;
  refresh.textContent = "Scanning…";
  status.className = "status scanning";
  status.textContent = "Looking for Chromecast devices…";

  try {
    const found = await window.go.main.App.DiscoverDevices();
    renderDevices(found || []);
    status.className = "status";
    status.textContent = `${found?.length || 0} device${found?.length === 1 ? "" : "s"} found`;
  } catch (error) {
    devices.innerHTML = "";
    status.className = "status error";
    status.textContent = `Discovery failed: ${error}`;
  } finally {
    refresh.disabled = false;
    refresh.textContent = "Scan again";
  }
}

devices.addEventListener("click", (event) => {
  const deviceElement = event.target.closest("[data-device-index]");
  if (!deviceElement) {
    return;
  }
  const device = renderedDevices[Number(deviceElement.dataset.deviceIndex)];
  if (device) {
    openDevice(device);
  }
});

appsElement.addEventListener("click", (event) => {
  const appElement = event.target.closest("[data-app-id]");
  if (!appElement) {
    return;
  }
  selectedAppID = appElement.dataset.appId;
  renderApps(renderedApps);
});

appControls.addEventListener("click", (event) => {
  const button = event.target.closest("#app-action");
  if (button) {
    runAppAction(button);
  }
});

appControls.addEventListener("submit", (event) => {
  if (event.target.matches("#youtube-form")) {
    event.preventDefault();
    playYouTube(event.target);
  }
});

back.addEventListener("click", showDeviceList);
refresh.addEventListener("click", scan);
scan();
