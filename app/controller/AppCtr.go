package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/open_tool/app/logic/chatGpt"
	"github.com/open_tool/app/logic/keyPool"
	"net/http"
	"strings"
	"time"
)

type AppCtr struct {
	*BaseCtr
}

func (c AppCtr) IndexPage(g *gin.Context) {
	g.HTML(http.StatusOK, "index/index.html", gin.H{
		"title": "0x318",
	})
}

func (c AppCtr) KeyPoolPage(g *gin.Context) {
	g.HTML(http.StatusOK, "key_pool/index.html", gin.H{})
}

func (c AppCtr) MakeFakePool(g *gin.Context) {
	// 设置响应头
	g.Writer.Header().Set("Content-Type", "text/event-stream;charset=utf-8")

	// 接收参数 account
	account := g.Query("account")
	if account == "" {
		keyPool.FmtSse(g.Writer, map[string]interface{}{
			"type": "error",
			"msg":  "参数错误",
		})
		return
	}
	accountList := strings.Split(account, "\n")
	// 去除空行
	for i := 0; i < len(accountList); i++ {
		if accountList[i] == "" {
			accountList = append(accountList[:i], accountList[i+1:]...)
			i--
		}
	}

	// 开始处理
	keyPool.MakePoolKey(g.Writer, accountList)
	// 输出结束标志
	keyPool.FmtSse(g.Writer, map[string]interface{}{"type": "end", "msg": "end"})
}

func (c AppCtr) MjPromptIndex(g *gin.Context) {
	g.HTML(http.StatusOK, "mj_prompt/index.html", gin.H{})
}

func (c AppCtr) MjPromptMake(g *gin.Context) {
	//g.JSON(http.StatusOK, map[string]any{
	//	"code":    200,
	//	"content": res,
	//})
}

func (c AppCtr) Text2SqlIndex(g *gin.Context) {
	svc := chatGpt.New("1")
	res, _ := svc.ChatWithContext("hi")
	g.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": res,
	})
	//g.HTML(http.StatusOK, "text2sql/index.html", gin.H{})
}

func (c AppCtr) Text2SqlTrans(g *gin.Context) {
	// 接收参数
	var req struct {
		Text   string `json:"text"`
		UserId string `json:"user_id"`
	}

	err := g.BindJSON(&req)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	svc := chatGpt.New(req.UserId)
	res, _ := svc.ChatWithContext(req.Text)

	g.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": res,
	})
}

func (c AppCtr) Stream(g *gin.Context) {
	// 接收参数 account
	account := g.Query("account")
	if account == "" {
		g.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}
	//accountList := strings.Split(account, "\n")

	// 定义一个消息通道，用于接收生成器的消息
	messageChan := make(chan string)
	defer close(messageChan)
	go func() {
		for i := 1; i <= 30; i++ {
			// 往隧道中写入消息

			da := map[string]interface{}{
				"loaded": i,
				"total":  30,
				"type":   "progress",
			}
			jsonData, _ := json.Marshal(da)

			messageChan <- "data: " + string(jsonData) + "\n\n"
		}

		// 写入结果
		da := map[string]interface{}{
			"pk":   "pk-123123y78suhfufdszjf",
			"type": "result",
		}

		jsonData, _ := json.Marshal(da)
		messageChan <- "data: " + string(jsonData) + "\n\n"
	}()

	w := g.Writer

	// 关闭输出缓冲，使得每次写入的数据能够立即发送给客户端
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	f, _ := w.(http.Flusher)
	// 从隧道中读取消息并写入到响应中
	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, msg)
			f.Flush()
			time.Sleep(100 * time.Millisecond)
		case <-time.After(5 * time.Second):
			// 在10秒内没有接收到数据，则发送一条注释，避免浏览器超时

			da := map[string]any{"type": "end"}
			jsonData, _ := json.Marshal(da)
			fmt.Fprintf(w, "data: "+string(jsonData)+"\n\n")
			f.Flush()
			return
		}
	}
}

func (c AppCtr) Ping(g *gin.Context) {
	g.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "pong",
	})
}

func (c AppCtr) ObtainTicket(g *gin.Context) {
	msg := "<!-- 5b85169e596104d563405be36f2cbe40ac4a2310f24eb4001b5e23f5bc8db302bfb5c74591a83dd537e083ef91d8851db78fe106f2ef1d487ed78aa26f3e1d67aae7647a83504b25df133f086cf8f7ed1115fe3438a54ef47d85c826f13e49fa5b7c65415c1fb4f1bc245cd2a76e18b05788efebd2519fa21828fa40f785395a20ac9d1ad2edb29e367a917c600983d8059edf39da358ea0f62d8c20f151367f0ab4abff7c8409f8e2c3e806496b80bba35ee8afbd006d39405cffd004964d92bb2241c9bb65057d5b17c0837291a4c1f148f99e3103df0b1bf12f3ed9a924b150576f599e0dfa73134056e7d8ce0ba86718d9b13a041871df292c5e249c366e -->\n<ObtainTicketResponse><action>NONE</action><confirmationStamp>1687933425178:8670e204-9838-42e9-8634-dd683ee3e4ef:SHA1withRSA:IHtWRXEzw+uNfrmyikWxW9ecx/Qa6u2b70KzMEAfEkgAATbI0l3nsY2IQ1qFQ+UzaYnu7A3pZZOu9xp+s4x4xo+9PruG4Rt5vEBKzmTktLzIe5guwk1vpXQLQTpFC6PeecjcRoa5SlIDkcvO1KZzCcDpJDYfK3mjtrFkELo2I/9ANpZE/FRWfZfE54D3ZxGlU4L/oKwH8n9+zkci3JTOMzxOhk8hpA6jrVJJPG7Mp/TcvTtO0y4km+XfiesILy5HVpLH+25/TQ3+ypCHldCy3MOZx9vVuXZAMu7T2wm1G/wQetZimVFjN8HDGwy/rB1Dow/ze2YI+LQoOO9ZlgJvXQ==:MIIEeDCCAmCgAwIBAgIIeSl8l4AETw4wDQYJKoZIhvcNAQELBQAwHTEbMBkGA1UEAwwSTGljZW5zZSBTZXJ2ZXJzIENBMCAXDTIyMDgxMjE1NDIyMloYDzIwOTkxMjMwMTYwMDAwWjAlMSMwIQYDVQQDDBpuYXNsbGVyLmxzcnYuamV0YnJhaW5zLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKnwFfq8kFX+APvtOLmyn024OSyWltN9X3rs7/rL2IN5gexsjNtkUXPkNbjkE+lnnf74Zw1t755CwY8kBvGJSkivfErb+c1CBIdLRFTBSJ6/tfzrPewYtDBMDuZbtEygyG9B1SWDoXYKkfGZxqcAbiEa3I1Pc/RooK4c8VZGyT++IgsECtZoLMRStcIotiYG+VPRPvd6w9+zmBUTWrpGenT30fYVlFSYLtudSTfUoeAOqBDqNXz8Jx5+ftkGjaZAwzkBH5PKLYa+dA09aJL5ZozHQ6Dm8FkK1amFhztR0w8HdC6MxU7pep1mR8UKHHpDqQEKnlwx9eEsoqLVetKvyA8CAwEAAaOBsTCBrjAMBgNVHRMBAf8EAjAAMA4GA1UdDwEB/wQEAwIFoDAgBgNVHSUBAf8EFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwTQYDVR0jBEYwRIAUpsRiEz+uvh6TsQqurtwXMd4J8VGhIaQfMB0xGzAZBgNVBAMMEkxpY2Vuc2UgU2VydmVycyBDQYIJAMCrW9HV+hjZMB0GA1UdDgQWBBQl3Z5j3vFAWfl/6/VDwmIgYjAXtTANBgkqhkiG9w0BAQsFAAOCAgEAZJe1OxDPwWQ7/TBABvNwue3CfS7GhHqP34SfmgWe+utlzXSHAKQ8RValLdOdRrIHbSTqAKMakXyznNZz9JAlbSNYQf7dpxg9LD+f+BULn9suktoUV7JkYDLm4Y9rucj1r35z9AG2obYnaZxTecqv8jjKrMkQwCvzXsUokGyq0E7topG8hwohCBWfPutCuCtnDKliBPUq6ua3OCOllI2GOcgGtmLICaeD/91q3KZZBoRkjUFRR72LUDON7vFwr/eOdU4d2vnIs1oMWjQFyOkc4Mq3GKEx3tdtaZrvZvIrnCcMlig7iiMXhhqaGgOBCeVpJ2zSW5lDIKSofzEaC8tiSMI1ClS9X6T2Wq9C/hSyxpBGYS1E1ZuLn+fUmutOFAhyF7dFzttCv9QvMeo2jGLbOAQ2zyeq84zoTSU6li2puHI9DHD05tS5+9jwkEtKLsBX1z0Mwl0+JY6kVUfc3OYfVVAYBWHn3eW9cJVaGN/59ymJV2UnrSsWPTVOXGz4H3GxGWCtnDpoEm6aDNzd5Z7H5Zz4C84g82GAPbKINpmyQqhNJw6PgIyUVniEpypsjLJ3fmFgfUWJg94vcfVkICU3CBY8Hy/dJ0NwSZ4e9TC0OZhq/LQqSYh/oROvX2fOdtYG38xtHxAOYv1Dy9TrM76rBOz4hhOE2nu75lfA8boCEAM=</confirmationStamp><leaseSignature>SHA512withRSA-ag1iBO5s4MQbj4Mk4mmOFPYP1utRrrwfmasutpYRelSoMi/G3A2y2fR9F9uDEG7YQfFXteOmTIpFG6h8d9mQwX1mQnyNLmoKGHdLTpL3zC7MuBoUaKCv2+iNW7VlAMmTDZczwhLls6goJlLVXBi3Bwm7/HISU52jogZzCBBMAwcw/rRc6n92wtIrSdinuFwiVqtkz4Q3VXc/z4VGCTnGARloxkjTDhKL5GdhYhx/EKNksdxoF/ilVsNQ86bOhyHBY9xOghIGXObtLLE+Nq2ZE86UchYTu9xgu1zPxTYXQZ0hDfguvm1iloELKGBCzaeSCH11x0deViZGrFpJ063EkA==-MIIEWzCCAkOgAwIBAgIINJagMxj4QmcwDQYJKoZIhvcNAQELBQAwGDEWMBQGA1UEAwwNSmV0UHJvZmlsZSBDQTAgFw0yMjA4MTIxNTQxMzRaGA8yMDk5MTIzMDE2MDAwMFowEjEQMA4GA1UEAwwHTmFzbGxlcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMcfMBkb2infMU/w5MyzYoZouQ9ghCc73hKKVlZwIl/Z/rBWyR0q73Il+cwAYC3qziUEjq+wNM4JvtdKfXtohRyWg578vqu4pbDzhudhXRqitNK7FY9yAevShabtRCEWCmGvp3gE4gEl8rorwlAKCMcYZIP1ujceEQwDss0HiQ6WhyUQ3VjI/VcH3+yjN0PfIFfz7/waEtqdXoXOP6L12jAfRA4Dg/8UjJaxFLy5+1sfiMReo61dk94vvshIr4Zebn97SidQp4/5Yb9sBTLFiXm4byhWTadmxqYJa1WzBIqag6HSFV1SOdVgGzrNp7GR65aLHWUrIBf8Hzsipbh5mxkCAwEAAaOBrDCBqTAMBgNVHRMBAf8EAjAAMA4GA1UdDwEB/wQEAwIFoDAgBgNVHSUBAf8EFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwSAYDVR0jBEEwP4AUo562SGdCEjZBvW3gubSgUouX8bOhHKQaMBgxFjAUBgNVBAMMDUpldFByb2ZpbGUgQ0GCCQDSbLGDsoN54TAdBgNVHQ4EFgQUM3Gc1UMiRTvhB7NrJzTB2i0QIpIwDQYJKoZIhvcNAQELBQADggIBAE3rgtdbxFiPN1KtXTVW7vDHJmyo9YFiNCMRLoQiwHC4qpM2oAU1qj+NzSyQ01t4dMR4gEpETlaYmrVKpwL1ltH0zA1jV/c/gI5PNJmVNbO2ylPHWoY23pVt3xVCQii657iimbdxRABrjYWeBGviI/0t8XkYDx2kAqEDYQ4DCyQRM2PFXVWCTsWE5a4CPVfMzvhJc60AO/PDuvSHyBFfF43Z47c3KS0EdCk+97Tj+mtImBQa1o4KRllVPI3B27arWAln+JrC1mKDtNluFtGBDl1vv+qcJSbnaPKnczzPCS06K2dalkEMFDoux4Xn61OZ683b2+s+56J6GpoiREEy3rWhXq5VHFOgrcsRi9jNWmcc1LnbK8iYnxGkGif4KtvMIsoAk4AzStpKTAotFYWvRyY6mxq02eBjUzkN4WAKnhHPz59jt/dH9kqkN0gtnVPPb45+bRXkCzOrwe9JUvv0KiH47Cg90L1oQc5Gq7kp2eWaN2cEycGMGfAziGAjSQRxFxGKMax/ROTJdGH29g+Y9PO1qXi1GsThsZ3yQgvU1Oh56TeUtPuR4tSOkK61N33zucLFWHrBGP6c3msG9PmNVZ/KGsJpAdTsDaQ/WBb30vV1/cP30Lgp22HNyzDEL+6LZKLCu8Ir8E6Pnj09QkICPHZqX8RujMPqT4r9M0/aVCUE</leaseSignature><message></message><prolongationPeriod>600000</prolongationPeriod><responseCode>OK</responseCode><salt>1687937092387</salt><serverLease>4102415999000:lew</serverLease><serverUid>lew</serverUid><ticketId>5hsf5j3jk1</ticketId><ticketProperties>licensee=lew licenseeType=5 metadata=0120211231PSAN000005</ticketProperties><validationDeadlinePeriod>-1</validationDeadlinePeriod><validationPeriod>600000</validationPeriod></ObtainTicketResponse>"
	// 页面输出msg 信息

	g.XML(200, gin.H{"msg": msg})
}
