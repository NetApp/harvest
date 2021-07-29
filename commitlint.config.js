module.exports = {
    extends: ['@commitlint/config-conventional'],
    rules: {
        'body-max-line-length': [0, 'always'],
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
