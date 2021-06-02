/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package prometheus

var html_template = `
<!DOCTYPE html>
<html>
    <head>
    <meta charset="utf-8">
    <title>NetApp Harvest 2.0 -%s- Prometheus Exporter</title>
    </head>
    <body>
    <br/>
    <h2 style="color:#404040">NetApp Harvest 2.0 - %s</h2>
    <p style="color:#303030">
    Welcome to the Prometheus Exporter of poller <em>%s</em>!<br/>
    If you are Prometheus scraper, get the metric data <a href="/metrics">here</a>.<br/><br/>

    Below is the list of metrics provided by my collectors and plugins.<br/>
    Exposing data from %d collectors and %d objects, %d metrics in total.<br/><br/>
    Note: this is a real-time generated list and might change over time.<br/>
    If you just started Harvest, you might need to wait a few minutes<br/>
    before the full list of counters is available here.<br/><br/>
    </p>
    %s
    </body>
</html>`

var body_template = `
        <div style="margin-left:40px; color:#303030">
            %s
        </div>`

var collector_template = `
            <h3 style="color:#404040">%s</h3>
            <small><em>collector</em></small>
            <ul>
                %s
            </ul>
            `

var object_template = `
                <h4 style="color:#404040">%s</h4>
                <small><em>object</em></small>
                %s`

var metric_template = `                <li>%s</li>`
