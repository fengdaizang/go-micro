package handler

//
//import (
//	"fmt"
//	"github.com/docker/docker/daemon/logger"
//	"github.com/gin-gonic/gin"
//	"io/ioutil"
//	"net/http"
//	"strings"
//	"tanghu.com/go-micro/common/starter/server/controller"
//)
//
//
////初始化
//func init() {
//	if err := controller.RegisterCtrl(&EventHandler{}); err != nil {
//		fmt.Println(err)
//	}
//}
//
//type EventHandler struct {
//}
//
////上传存证-合约网关回调函数
//func (eh *EventHandler) Hello(c *gin.Context) {
//	var req contract_gateway.TxContractRspTransactionInfo
//
//	//(1)读取请求体数据
//	body, _ := ioutil.ReadAll(c.Request.Body)
//
//	if err := json.Unmarshal(body, &req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
//	//(2)根据txId获取数据库数据
//	info, err := models.QueryEvidencePropertyInfoByTxId(req.TxId)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	//(3)根据validateCode更新数据库状态
//	if strings.EqualFold(req.ValidationCode, contract_gateway.VALID) {
//		info.State = domain.CreatedState
//		info.BlockNumber = uint64(req.BlockNumber)
//	} else {
//		info.State = domain.CreateFailedState
//	}
//
//	_, err = models.UpdateEvidencePropertyInfo(info)
//	if err != nil {
//		logger.Error("UpdateEvidencePropertyInfo failed when receiptEvent:", err.Error())
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	return
//}
//
////删除存证-合约网关回调函数
//func (eh *EventHandler) CallbackDeleteEvent(c *gin.Context) {
//	var req contract_gateway.TxContractRspTransactionInfo
//
//	//(1)读取请求体数据
//	body, _ := ioutil.ReadAll(c.Request.Body)
//
//	if err := json.Unmarshal(body, &req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
//	//(2)根据txId获取数据库数据
//	info, err := models.QueryEvidencePropertyInfoByTxId(req.TxId)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	//(3)根据validateCode更新数据库状态
//	if strings.EqualFold(req.ValidationCode, contract_gateway.VALID) {
//		info.State = domain.DeletedState
//		info.Deleted = true
//		info.BlockNumber = uint64(req.BlockNumber)
//	} else {
//		info.State = domain.DeleteFailedState
//	}
//
//	_, err = models.UpdateEvidencePropertyInfo(info)
//	if err != nil {
//		logger.Error("UpdateEvidencePropertyInfo failed when receiptEvent:", err.Error())
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	return
//}
//
////返回名称
//func (eh *EventHandler) Name() string {
//	return "event"
//}
//
////配置路由
//func (eh *EventHandler) Routers() []controller.Router {
//	return []controller.Router{
//		{
//			Path:    "/callback/upload",
//			Method:  controller.POST,
//			Handler: eh.CallbackUploadEvent,
//		},
//		{
//			Path:    "/callback/delete",
//			Method:  controller.POST,
//			Handler: eh.CallbackDeleteEvent,
//		},
//	}
//}
//
////参数校验
//func (eh *EventHandler) Middlewares() []string {
//	return []string{}
//}
//
