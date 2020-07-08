package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"sort"
	"strconv"
	"strings"
	"vmapi/model"
	"vmapi/module"
)

// @Summary Get ip pool
// @Description Get ip pool
// @Tags IPS
// @Accept json
// @Produce json
// @Success 200 {object} model.IPS
// @Failure 500 {object} model.Err
// @Router /ips [get]
func IPS(ctx *gin.Context) {
	var ipool model.IPS
	if val, err := module.RedisGet("ipool"); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	} else {
		if err := json.Unmarshal([]byte(val), &ipool); err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: err,
			})
			return
		}
	}
	ctx.JSON(200, ipool)
}

// @Summary Create ip pool
// @Description Create ip pool
// @Tags IPS
// @Accept json
// @Produce json
// @Param net body model.NET true "IP地址段"
// @Success 200 {object} model.Res
// @Failure 500 {object} model.Err
// @Router /ips [post]
func IPSCreate(ctx *gin.Context) {
	var ips model.NET
	if err := ctx.ShouldBindJSON(&ips); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	var ipool model.IPS
	field := strings.Split(ips.NETWORK, "-")
	sp := strings.Split(field[0], ".")
	net := strings.Join([]string{sp[0], sp[1], sp[2]}, ".")
	s, _ := strconv.Atoi(sp[3])
	e, _ := strconv.Atoi(field[1])
	for i := s; i <= e; i++ {
		ipool = append(ipool, net+"."+strconv.Itoa(i))
	}
	data, err := json.Marshal(ipool)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	if err := module.RedisSet("ipool", string(data)); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	ctx.JSON(200, model.Res{
		Error:   200,
		Message: "ip pool create ok!",
	})
}

// @Summary Remove a ip from ip pool
// @Description Remove a ip from ip pool
// @Tags IPS
// @Accept json
// @Produce json
// @Param ip path string true "IP"
// @Success 200 {object} model.Res
// @Failure 500 {object} model.Err
// @Router /ips/{ip} [delete]
func IPRemove(ctx *gin.Context) {
	ip := ctx.Param("ip")
	var ipool model.IPS
	if val, err := module.RedisGet("ipool"); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	} else {
		if err := json.Unmarshal([]byte(val), &ipool); err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: err,
			})
			return
		}
	}
	// 二分法查找
	sort.Strings(ipool)
	index := sort.Search(len(ipool), func(i int) bool {
		return ipool[i] >= ip
	})

	ipool = append(ipool[:index], ipool[index+1:]...)
	data, err := json.Marshal(ipool)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}
	//写回redis
	if err := module.RedisSet("ipool", string(data)); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	ctx.JSON(200, model.Res{
		Error:   200,
		Message: fmt.Sprintf("%s remove pool ok!", ip),
	})
}


// @Summary Append a ip to ip pool
// @Description Append a ip to ip pool
// @Tags IPS
// @Accept json
// @Produce json
// @Param ip path string true "IP"
// @Success 200 {object} model.Res
// @Failure 500 {object} model.Err
// @Router /ips/{ip} [put]
func IPAppend(ctx *gin.Context) {
	ip := ctx.Param("ip")
	var ipool model.IPS
	if val, err := module.RedisGet("ipool"); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	} else {
		if err := json.Unmarshal([]byte(val), &ipool); err != nil {
			ctx.JSON(500, model.Err{
				Error:   500,
				Message: err,
			})
			return
		}
	}

	ipool = append(ipool, ip)
	data, err := json.Marshal(ipool)
	if err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	//写回redis
	if err := module.RedisSet("ipool", string(data)); err != nil {
		ctx.JSON(500, model.Err{
			Error:   500,
			Message: err,
		})
		return
	}

	ctx.JSON(200, model.Res{
		Error:   200,
		Message: fmt.Sprintf("%s append pool ok!", ip),
	})
}