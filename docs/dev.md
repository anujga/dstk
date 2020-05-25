  
# Dev Guidlines


### Many of the methods are marked as path=control/data.

They should be handled as

1. Performance: Not critical in control

1. Logging: For control plane
    1. Use sugared logger.
    1. Prefer log level=INFO in order to form an audit log. If a running
    setup suddenly degrades, it is likely to coincide with
    one of the control operations. Emit relevant state wrt the control
    operation. For data plane stick to metrics for hot path. For
    exceptional we don't care about performance but so many logs during
    some failure will stress the system unnecessarily. Use throttled
    logs or metrics.
             
