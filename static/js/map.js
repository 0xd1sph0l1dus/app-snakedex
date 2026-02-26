function initObsMap() {
  var el = document.getElementById('obs-map');
  if (!el || el._leaflet_id) return; // already initialised or absent

  var raw = el.getAttribute('data-coords');
  if (!raw) return;

  var coords;
  try { coords = JSON.parse(raw); } catch (e) { return; }
  if (!coords.length) return;

  var map = L.map('obs-map', { zoomControl: true });
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
  }).addTo(map);

  var group = L.featureGroup();
  coords.forEach(function(c) {
    L.circleMarker(c, {
      radius: 6,
      color: '#5aff6e',
      fillColor: '#5aff6e',
      fillOpacity: 0.65,
      weight: 1.5
    }).addTo(group);
  });
  group.addTo(map);

  if (group.getBounds().isValid()) {
    map.fitBounds(group.getBounds().pad(0.25));
  }
}

// Full-page: init on load.
document.addEventListener('DOMContentLoaded', initObsMap);

// HTMX modal: init after content settles (modal is visible by then).
document.addEventListener('htmx:afterSettle', initObsMap);
