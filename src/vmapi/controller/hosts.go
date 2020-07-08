package controller

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"vmapi/model"
	"vmapi/module"
	"vmapi/templates"
)

// @Summary Host Monitor
// @Description Host Monitor
// @Tags HOSTS
// @Accept json
// @Produce json
// @Param host path string true "HOST"
// @Success 200 {object} model.Res
// @Failure 500 {object} model.Err
// @Router /host/{host} [get]
func HOSTDetail(ctx *gin.Context) {
	phy := ctx.Param("host")
	if err := module.SFTPut(phy, "root", "dd@2019", 22, templates.Script,
		"/tmp/sys.py"); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	sshClient := module.SSHConnect(phy, "root", "dd@2019", 22)
	defer sshClient.Close()
	// 创建ssh-session
	session, _ := sshClient.NewSession()
	defer session.Close()

	var stdOut, stdErr bytes.Buffer
	session.Stdout = &stdOut
	session.Stderr = &stdErr

	session.Run("python /tmp/sys.py")
	var sys model.SYS
	json.Unmarshal(stdOut.Bytes(), &sys)
	ctx.JSON(200, sys)
}

// @Summary Host VM Detail
// @Description Host VM Detail
// @Tags HOSTS
// @Accept json
// @Produce json
// @Param host path string true "HOST"
// @Success 200 {object} model.Res
// @Failure 500 {object} model.Err
// @Router /hosts/vm/{host} [get]
func HOSTVMDetail(ctx *gin.Context) {
	phy := ctx.Param("host")
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

	var vmsdetail model.HOSTVM
	for name, state := range allDoms {
		domInfo, _ := conn.LookupDomainByName(name)
		vmInfo, _ := module.RedisGet(name)
		vmData := strings.Split(vmInfo, "::")
		vmsdetail = append(vmsdetail, model.VMC{
			ID:     func() uint { id, _ := domInfo.GetID(); return id }(),
			NAME:   name,
			UUID:   func() string { uuid, _ := domInfo.GetUUIDString(); return uuid }(),
			CPU:    func() int { cpu, _ := strconv.Atoi(vmData[5]); return cpu }(),
			MEMORY: vmData[6],
			STATUS: state,
			NETWORK: model.NETWORK{
				IP:      vmData[1],
				NETMASK: vmData[2],
				GATEWAY: vmData[3],
				DNS: []string{
					"114.114.114.114",
					"223.5.5.5",
				},
			},
		})
	}

	ctx.JSON(200, vmsdetail)
}