module.exports = {
    extends: ['@commitlint/config-conventional'],
    rules: {
        'scope-enum': [
            2,
            'always',
            [
                'collector',
                'config',
                'exporter',
                'grafana',
                'influxdb',
                'manager',
                'matrix',
                'poller',
                'prometheus',
            ]
        ]
    }
};


