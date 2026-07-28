const devices = document.querySelector("#devices");
const status = document.querySelector("#status");
const refresh = document.querySelector("#refresh");

function escapeHTML(value) {
  const node = document.createElement("span");
  node.textContent = value || "";
  return node.innerHTML;
}

function render(found) {
  if (!found.length) {
    devices.innerHTML = `
      <div class="empty">
        <strong>No Cast devices found</strong>
        <span>Check that this computer and your Chromecast are on the same network.</span>
      </div>`;
    return;
  }

  devices.innerHTML = found.map((device) => `
    <article class="device">
      <div class="device-icon" aria-hidden="true">◧</div>
      <div class="device-copy">
        <h2>${escapeHTML(device.name)}</h2>
        <p>${escapeHTML(device.model || "Google Cast device")}</p>
        <code>${escapeHTML(device.host)}:${device.port}</code>
      </div>
      <span class="state">${escapeHTML(device.connectionState)}</span>
    </article>`).join("");
}

async function scan() {
  refresh.disabled = true;
  refresh.textContent = "Scanning…";
  status.className = "status scanning";
  status.textContent = "Looking for Chromecast devices…";

  try {
    const found = await window.go.main.App.DiscoverDevices();
    render(found || []);
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

refresh.addEventListener("click", scan);
scan();
