const GRAFANA_HOST = "grafanaHost";
const SLASH_D = "/d/";

document$.subscribe(function() {

    document.addEventListener('click', function (event) {
        const link = event.target.closest('.grafana-table a');
        if (!link) return;

        const index = link.href.indexOf(SLASH_D);

        if (index === -1) return;

        event.preventDefault();

        const grafanaHost = localStorage.getItem(GRAFANA_HOST);

        if (!grafanaHost) {
            alert('Please enter a Grafana host URL to open links to your dashboards.');
            return;
        }

        const dashboard = link.href.substring(index + 1);

        window.open(grafanaHost + dashboard, '_blank');
    });

    const grafanaHostInput = document.getElementById(GRAFANA_HOST);
    if (!grafanaHostInput) return;

    grafanaHostInput.addEventListener('keydown', function(event) {
        if (event.key === 'Enter') {
            event.preventDefault();
            saveGrafanaHost();
        }
    });
});

function isValidUrl(grafanaHost) {
    try {
        new URL(grafanaHost);
        return true;
    } catch (e) {
        return false;
    }
}

function saveGrafanaHost() {
    const grafanaHostInput = document.getElementById(GRAFANA_HOST);
    if (!grafanaHostInput) {
        alert('Grafana host input not found.');
        return;
    }

    let grafanaHost = grafanaHostInput.value.trim();
    if (!grafanaHost) {
        alert('Please enter a Grafana host URL.');
        return;
    }

    if (!isValidUrl(grafanaHost)) {
        alert('Please enter a valid Grafana host URL.');
        return;
    }

    if (!grafanaHost.endsWith('/')) {
        grafanaHost += '/';
    }
    localStorage.setItem(GRAFANA_HOST, grafanaHost);
}