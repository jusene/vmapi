package controller

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"libvirt.org/libvirt-go"
	"path/filepath"
	"strconv"
	"strings"
	"vmapi/model"
	"vmapi/module"
	"vmapi/templates"
)

// @Summary Get all vms
// @Description Get all vms
// @Tags VMS
// @Accept json
// @Produce json
// @Param phy path string true "物理机IP"
// @Success 200 {object} model.VMS
// @Failure 500 {object} model.Err
// @Router /vms/{phy} [get]
func VMList(ctx *gin.Context) {
	phy := ctx.Param("phy")
	conn, err := module.LibvirtConnect(phy)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}
	defer conn.Close()
	runDoms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}
	shutDoms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	vms := model.VMS{}
	for _, dom := range runDoms {
		name, err := dom.GetName()
		if err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: err,
			})
			return
		}
		vms = append(vms, model.VM{Name: name, State: "running"})
	}

	for _, dom := range shutDoms {
		name, err := dom.GetName()
		if err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: err,
			})
			return
		}
		vms = append(vms, model.VM{Name: name, State: "shut off"})
	}
	ctx.JSON(200, vms)
}

// @Summary Create a vm
// @Description Create a vm
// @Tags VMS
// @Accept json
// @Produce json
// @Param vm body model.VMDetail true "vm"
// @Success 200 {object} model.Res
// @Success 201 {object} model.Res
// @Failure 500 {object} model.Err
// @Router /vms [post]
func VMCreate(ctx *gin.Context) {
	var vm model.VMDetail
	if err := ctx.ShouldBindJSON(&vm); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}
	// 解析模板
	temp, err := template.ParseFiles("templates/VMXMLDesc.xml")
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}
	kvm := new(model.VMDetail)
	kvm.NAME = vm.NAME
	kvm.CPU = vm.CPU
	kvm.MEMORY = vm.MEMORY
	kvm.IMAGE = fmt.Sprintf("/ddhome/kvm/images/%s.qcow2", vm.NAME)
	var buf bytes.Buffer
	if err := temp.ExecuteTemplate(&buf, "kvmxml", &kvm); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}
	conn, err := module.LibvirtConnect(vm.PhyIP)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}
	// 创建镜像并定义虚拟机
	if err = module.SSHExec(vm.PhyIP, "root", "dd@2019", 22,
		fmt.Sprintf("/usr/bin/qemu-img create -f qcow2 -b %s %s",
			strings.Join([]string{filepath.Dir(kvm.IMAGE), "model.qcow2"}, "/"), kvm.IMAGE)); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
	}

	if _, err = conn.DomainDefineXML(buf.String()); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}
	// 设置虚拟机网络
	if err := kvmNet(&vm); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	// 启动虚拟机
	dom, err := conn.LookupDomainByName(vm.NAME)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	if err = dom.Create(); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	// 记录虚拟机信息
	vmList := strings.Join([]string{vm.NAME, vm.IPADDR, vm.NETMASK, vm.GATEWAY, vm.PhyIP, vm.CPU, vm.MEMORY}, "::")
	if err = module.RedisSet(vm.NAME, vmList); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	ctx.JSON(201, model.Res{
		Error:   201,
		Message: fmt.Sprintf("%s create ok", vm.NAME),
	})
}

func kvmNet(vm *model.VMDetail) error {
	// 网络设置
	temp, err := template.New("net").Parse(templates.NET)
	if err != nil {
		return err
	}
	net := new(model.VMDetail)
	net.IPADDR = vm.IPADDR
	net.NETMASK = vm.NETMASK
	net.GATEWAY = vm.GATEWAY
	var buf bytes.Buffer
	if err := temp.Execute(&buf, net); err != nil {
		return err
	}
	if err = module.SSHExec(vm.PhyIP, "root", "dd@2019", 22,
		fmt.Sprintf("mkdir -p /ddhome/kvm/config/%s", vm.NAME)); err != nil {
		return err
	}
	if err = module.SFTPut(vm.PhyIP, "root", "dd@2019", 22, buf.String(),
		fmt.Sprintf("/ddhome/kvm/config/%s/ifcfg-eth0", vm.NAME)); err != nil {
		return err
	}
	if err = module.SSHExec(vm.PhyIP, "root", "dd@2019", 22,
		fmt.Sprintf("/usr/bin/virt-copy-in -d %s /ddhome/kvm/config/%s/ifcfg-eth0 /etc/sysconfig/network-scripts", vm.NAME, vm.NAME)); err != nil {
		return err
	}
	return nil
}

// @Summary Get a vm detail
// @Description Get a vm detail
// @Tags VMS
// @Accept json
// @Produce json
// @Param phy path string true "物理机IP"
// @Param vm path string true "虚拟机NAME"
// @Success 200 {object} model.VMC
// @Success 404 {object} model.Err
// @Failure 500 {object} model.Err
// @Router /vms/{phy}/{vm} [get]
func VMDetail(ctx *gin.Context) {
	phy := ctx.Param("phy")
	vm := ctx.Param("vm")
	conn, err := module.LibvirtConnect(phy)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	allDoms, err := GetAllDom(conn)()
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	if _, ok := allDoms[vm]; !ok {
		ctx.JSON(404, model.Err{
			Error:   404,
			Message: fmt.Sprintf("%s Not Found", vm),
		})
		return
	}

	// 查找虚拟机基本信息
	vmInfo, err := module.RedisGet(vm)
	if err != nil {
		ctx.JSON(404, model.Err{
			Error:   404,
			Message: err,
		})
		return
	}

	vmData := strings.Split(vmInfo, "::")

	domInfo, err := conn.LookupDomainByName(vm)
	detail := model.VMC{
		ID:     func() uint { id, _ := domInfo.GetID(); return id }(),
		NAME:   func() string { name, _ := domInfo.GetName(); return name }(),
		UUID:   func() string { uuid, _ := domInfo.GetUUIDString(); return uuid }(),
		CPU:    func() int { cpu, _ := strconv.Atoi(vmData[5]); return cpu }(),
		MEMORY: vmData[6],
		STATUS: func() string {
			info, _ := domInfo.GetInfo()
			state := map[libvirt.DomainState]string{
				libvirt.DOMAIN_NOSTATE:     "nostate",
				libvirt.DOMAIN_RUNNING:     "running",
				libvirt.DOMAIN_BLOCKED:     "blocked",
				libvirt.DOMAIN_PAUSED:      "paused",
				libvirt.DOMAIN_SHUTDOWN:    "shutdown",
				libvirt.DOMAIN_CRASHED:     "crashed",
				libvirt.DOMAIN_PMSUSPENDED: "pmsuspended",
				libvirt.DOMAIN_SHUTOFF:     "shutoff",
			}
			return state[info.State]
		}(),
		NETWORK: model.NETWORK{
			IP:      vmData[1],
			NETMASK: vmData[2],
			GATEWAY: vmData[3],
			DNS: []string{
				"114.114.114.114",
				"223.5.5.5",
			},
		},
	}
	ctx.JSON(200, detail)
}

func GetAllDom(conn *libvirt.Connect) func() (map[string]string, error) {
	domMap := make(map[string]string, 20)
	return func() (map[string]string, error) {
		defer conn.Close()
		runDoms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
		if err != nil {
			return nil, err
		}
		shutDoms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
		if err != nil {
			return nil, err
		}
		for _, dom := range runDoms {
			name, _ := dom.GetName()
			domMap[name] = "running"
		}

		for _, dom := range shutDoms {
			name, _ := dom.GetName()
			domMap[name] = "shutoff"
		}
		return domMap, nil
	}
}

// @Summary Shutdown a vm
// @Description Shutdown a vm
// @Tags VMS
// @Accept json
// @Produce json
// @Param Force header bool false "强制关机"
// @Param phy path string true "物理机IP"
// @Param vm path string true "虚拟机NAME"
// @Success 200 {object} model.Res
// @Failure 403 {object} model.Err
// @Failure 500 {object} model.Err
// @Router /vms/{phy}/{vm} [delete]
func VMDelete(ctx *gin.Context) {
	phy := ctx.Param("phy")
	vm := ctx.Param("vm")
	force := ctx.GetHeader("Force")
	conn, err := module.LibvirtConnect(phy)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	allDoms, err := GetAllDom(conn)()
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	if _, ok := allDoms[vm]; !ok {
		ctx.JSON(404, model.Err{
			Error:   404,
			Message: fmt.Sprintf("%s Not Found", vm),
		})
		return
	}

	domInfo, err := conn.LookupDomainByName(vm)
	info, _ := domInfo.GetInfo()
	if info.State == libvirt.DOMAIN_SHUTOFF {
		ctx.JSON(403, model.Err{
			Error:   403,
			Message: fmt.Sprintf("%s aleardy shutoff!", vm),
		})
		return
	} else if info.State == libvirt.DOMAIN_RUNNING && func() bool {
		b, _ := strconv.ParseBool(force)
		return b
	}() {
		if err := domInfo.Destroy(); err != nil {
			ctx.JSON(500, model.Res{
				Error:   500,
				Message: fmt.Sprintf("%s force shutoff error!", vm),
			})
			return
		}
		ctx.JSON(200, model.Res{
			Error:   200,
			Message: fmt.Sprintf("%s force shutoff ok!", vm),
		})
		return
	} else if info.State == libvirt.DOMAIN_RUNNING {
		if err := domInfo.Shutdown(); err != nil {
			ctx.JSON(500, model.Res{
				Error:   500,
				Message: fmt.Sprintf("%s shutoff error!", vm),
			})
			return
		}
		ctx.JSON(200, model.Res{
			Error:   200,
			Message: fmt.Sprintf("%s shutoff ok!", vm),
		})
		return
	}
}

// @Summary Controller a vm
// @Description Controller a vm
// @Tags VMS
// @Accept json
// @Produce json
// @Param operator body model.OP false "操作"
// @Param phy path string true "物理机IP"
// @Param vm path string true "虚拟机NAME"
// @Success 200 {object} model.Res
// @Failure 500 {object} model.Err
// @Router /vms/{phy}/{vm} [put]
func VMController(ctx *gin.Context) {
	phy := ctx.Param("phy")
	vm := ctx.Param("vm")
	var op model.OP
	ctx.ShouldBindJSON(&op)
	fmt.Println("OPERATOR", op.OPERATOR)
	conn, err := module.LibvirtConnect(phy)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	allDoms, err := GetAllDom(conn)()
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	if _, ok := allDoms[vm]; !ok {
		ctx.JSON(404, model.Err{
			Error:   404,
			Message: fmt.Sprintf("%s Not Found", vm),
		})
		return
	}

	domInfo, err := conn.LookupDomainByName(vm)
	info, _ := domInfo.GetInfo()

	// 重启虚拟机
	if op.OPERATOR == "reboot" && info.State == libvirt.DOMAIN_RUNNING {
		if err := domInfo.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT); err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: fmt.Sprintf("%s reboot error", vm),
			})
			return
		}
		ctx.JSON(200, model.Res{
			Error:   200,
			Message: fmt.Sprintf("%s reboot ok!", vm),
		})
		return
	} else if op.OPERATOR == "reboot" && info.State == libvirt.DOMAIN_SHUTOFF {
		ctx.JSON(403, model.Err{
			Error:   403,
			Message: fmt.Sprintf("%s aleardy shutoff!", vm),
		})
		return
	}

	// 销毁机器
	if op.OPERATOR == "delete" && info.State == libvirt.DOMAIN_RUNNING {
		if err := domInfo.Destroy(); err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: fmt.Sprintf("%s destroy error", vm),
			})
			return
		}
	}

	if op.OPERATOR == "delete" {
		if err := domInfo.Undefine(); err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: fmt.Sprintf("%s undefine error", vm),
			})
			return
		}

		if err = module.SSHExec(phy, "root", "dd@2019", 22,
			fmt.Sprintf("rm -rf /ddhome/kvm/images/%s.qcow2", vm)); err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: fmt.Sprintf("delete %s.qcow2 error", vm),
			})
			return
		}

		if err = module.SSHExec(phy, "root", "dd@2019", 22,
			fmt.Sprintf("rm -rf /ddhome/kvm/config/%s", vm)); err != nil {
			{
				ctx.JSON(500, model.Err{
					Error:   500,
					Message: "delete network config error",
				})
				return
			}
		}

		ctx.JSON(200, model.Err{
			Error:   200,
			Message: fmt.Sprintf("delete %s ok", vm),
		})
	}

	// 开启操作
	if op.OPERATOR == "" || op.OPERATOR == "string" {
		if err := domInfo.Create(); err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: fmt.Sprintf("%s start error", vm),
			})
			return
		}
		ctx.JSON(200, model.Err{
			Error:   200,
			Message: fmt.Sprintf("%s start ok", vm),
		})
	}
}
