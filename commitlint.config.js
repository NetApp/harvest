module.exports = {
    extends: ['@commitlint/config-conventional'],
    rules: {
        // Disable the following ones
        'body-max-line-length': [0, 'always'],
        'subject-case': [0, 'always'],
        'footer-max-line-length': [0, 'always'],
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
