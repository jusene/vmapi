## 部署文档

在改接口程序运行的服务器安装：
```bash
yum install -y epel-release
yum install -y libvirt redis
```

启动redis
```bash
systemctl enable redis --now
```

在程序运行的目录下必须有templates目录，里面需要保存VMXMLDesc.xml文件
example:
```xml
{{ define "kvmxml"}}
<domain type='kvm'>
    <name>{{ .NAME }}</name>
    <memory unit='KiB'>{{ .MEMORY }}</memory>
    <currentMemory unit='KiB'>{{ .MEMORY }}</currentMemory>
    <vcpu placement='static'>{{ .CPU }}</vcpu>
    <os>
        <type arch='x86_64' machine='pc-i440fx-rhel7.0.0'>hvm</type>
        <boot dev='hd'/>
    </os>
    <features>
        <acpi/>
        <apic/>
    </features>
    <!--
    <cpu mode='custom' match='exact' check='partial'>
      <model fallback='allow'>Skylake-Client-IBRS</model>
      <feature policy='require' name='spec-ctrl'/>
      <feature policy='require' name='ssbd'/>
    </cpu>
    -->
    <cpu mode='custom' match='exact' check='full'>
        <model fallback='forbid'>Broadwell</model>
        <feature policy='require' name='hypervisor'/>
        <feature policy='require' name='xsaveopt'/>
    </cpu>
    <clock offset='utc'>
        <timer name='rtc' tickpolicy='catchup'/>
        <timer name='pit' tickpolicy='delay'/>
        <timer name='hpet' present='no'/>
    </clock>
    <on_poweroff>destroy</on_poweroff>
    <on_reboot>restart</on_reboot>
    <on_crash>destroy</on_crash>
    <pm>
        <suspend-to-mem enabled='no'/>
        <suspend-to-disk enabled='no'/>
    </pm>
    <devices>
        <emulator>/usr/libexec/qemu-kvm</emulator>
        <disk type='file' device='disk'>
            <driver name='qemu' type='qcow2'/>
            <source file='{{ .IMAGE }}'/>
            <target dev='vda' bus='virtio'/>
            <address type='pci' domain='0x0000' bus='0x00' slot='0x07' function='0x0'/>
        </disk>
        <controller type='pci' index='0' model='pci-root'/>
        <controller type='ide' index='0'>
            <address type='pci' domain='0x0000' bus='0x00' slot='0x01' function='0x1'/>
        </controller>
        <controller type='virtio-serial' index='0'>
            <address type='pci' domain='0x0000' bus='0x00' slot='0x06' function='0x0'/>
        </controller>
        <interface type='bridge'>
            <source bridge='br0'/>
            <model type='virtio'/>
            <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x0'/>
        </interface>
        <serial type='pty'>
            <target type='isa-serial' port='0'>
                <model name='isa-serial'/>
            </target>
        </serial>
        <console type='pty'>
            <target type='serial' port='0'/>
        </console>
        <channel type='unix'>
            <target type='virtio' name='org.qemu.guest_agent.0'/>
            <address type='virtio-serial' controller='0' bus='0' port='1'/>
        </channel>
        <channel type='spicevmc'>
            <target type='virtio' name='com.redhat.spice.0'/>
            <address type='virtio-serial' controller='0' bus='0' port='2'/>
        </channel>
        <input type='mouse' bus='ps2'/>
        <input type='keyboard' bus='ps2'/>
        <graphics type='spice' autoport='yes'>
            <listen type='address'/>
            <image compression='off'/>
        </graphics>
        <memballoon model='virtio'>
            <address type='pci' domain='0x0000' bus='0x00' slot='0x08' function='0x0'/>
        </memballoon>
    </devices>
</domain>
{{ end }}
```

建立地址池
地址：http://IP:8098/swagger/index.html
格式 {"network": "192.168.66.220-240"}

在运行虚拟机的物理服务器上，安装kvm
```bash
yum install -y libvirt virt-manager virt-viewer virt-install qemu-kvm  libguestfs-tools
```
- 修改配置/etc/libvirt/libvirtd.conf
```bash
listen_tls = 0
listen_tcp = 1
tcp_port = "16509"
listen_addr = "0.0.0.0"
auth_tcp = "none"
```

- 修改配置/etc/sysconfig/libvirtd
```bash
LIBVIRTD_ARGS="--listen -f /etc/libvirt/libvirtd.conf"
```

- 启动服务
```bash
systemctl start libvirtd
```

- 安装psutil
```bash
yum install -y epel-release python-pip python-devel
pip install psutil
```

- 准备目录
```
mkdir -p /ddhome/kvm/images

把做好的镜像放入这个目录  并且命名为model.qcow2
```



