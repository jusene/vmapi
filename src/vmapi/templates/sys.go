package templates

var Script = `
#!/usr/bin/env python

import psutil
import json

sys_info = {
    "cpu": {
        "count": psutil.cpu_count(),
        'cpu_percent': psutil.cpu_percent(interval=0.5)
    },
    "mem": dict(zip(["total", "available", "percent", "used", "free", "active", "inactive", "buffers", "cached", "shared", "slab"],
              list(psutil.virtual_memory()))),
    "disk": {"/ddhome": dict(zip(["total", "used", "free", "percent", 'mountpoint'], list(psutil.disk_usage("/ddhome"))))}
                                   

}

print(json.dumps(sys_info))
`
