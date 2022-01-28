# REST collector

## Status <!-- one of: In Progress, Accepted, Rejected, Superseded, Deprecated -->

Accepted

:construction: The exact version of ONTAP that has full ZAPI parity is subject to change. 
Everywhere you see version 9.12, may become 9.13 or later.

## Context

We need to document and communicate to customers:
- when they should switch from the ZAPI collectors to the REST ones
- what versions of ONTAP are supported by Harvest's REST collectors
- how to fill ONTAP gaps between the ZAPI and REST APIs

The ONTAP version information is important because gaps are addressed in later versions of cDOT.

## Considered Options

1. **Only REST** A clean cut-over, stop using ZAPI, and switch completely to REST.

2. **Both** Support both ZAPI and REST collectors running at the same time, collecting the same objects. Flexible, but has the downside of last-write wins. Not recommended unless you selectively pick non-overlapping sets of objects.

3. **Template change that supports both** Change the template to break ties, priority, etc. Rejected because additional complexity not worth the benefits.

4. **private-cli** When there are REST gaps that have not been filled yet or will never be filled (WONTFIX), the Harvest REST collector will provide infrastructure and documentation on how to use private-cli pass-through to address gaps.

## Chosen Decision

For clusters with ONTAP versions < 9.12, we recommend customers use the ZAPI collectors. (#2) (#4)

Once ONTAP 9.12+ is released and customers have upgraded to it, they should make a clean cut-over to the REST collectors (#1). 
ONTAP 9.12 is the version of ONTAP that has the best parity with what Harvest collects in terms of config and performance counters. 
Harvest REST collectors, templates, and dashboards are validated against ONTAP 9.12+. 
Most of the REST config templates will work before 9.12, but unless you have specific needs, we recommend sticking with the ZAPI collectors until you upgrade to 9.12.

There is little value in running both the ZAPI and REST collectors for an overlapping set of objects. 
It's unlikely you want to collect the same object via REST and ZAPI at the same time. Harvest doesn't support this use-case, but does nothing to detect or prevent it.

If you want to collect a non-overlapping set of objects with REST and ZAPI, you can. 
If you do, we recommend you disable the ZAPI object collector. 
For example, if you enable the REST `disk` template, you should disable the ZAPI `disk` template. 
We do NOT recommend collecting an overlapping set of objects with both collectors since the last one to run will overwrite previously collected data.

Harvest will document how to use the REST private cli pass-through to collect custom and non-public counters.

The Harvest team recommends that customers open ONTAP issues for REST public API gaps that need filled.

## Consequences

The Harvest REST collectors will work with limitations on earlier versions of ONTAP. 
ONTAP 9.12+ is the minimally validated version. 
We only validate the full set of templates, dashboards, counters, etc. on versions of ONTAP 9.12+

Harvest does not prevent you from collecting the same resource with ZAPI and REST.