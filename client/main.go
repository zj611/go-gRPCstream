package main

import (
	"client/stream-pb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"strconv"
)

func main() {
	serviceAddress := "127.0.0.1:1234"
	conn, err := grpc.Dial(serviceAddress, grpc.WithInsecure())
	if err != nil {
		panic("connect error")
	}
	defer conn.Close()
	stringClient := stream_pb.NewStringServiceClient(conn)
	stringReq := &stream_pb.StringRequest{A: "A", B: "B"}

	// 1.简单远程调用
	res, err := stringClient.Concat(context.Background(),stringReq)
	fmt.Println(res.Ret)

	// 2.流请求:单个客户端请求，服务端流式返回10个响应结果
	//stream, _ := stringClient.LotsOfServerStream(context.Background(), stringReq)
	//for {//流结果接收
	//	item, streamError := stream.Recv() //客户端流式接受响应结果
	//	if streamError == io.EOF {  //直到流结束符号终止
	//		break
	//	}
	//	if streamError != nil {//或者流出错的时候终止
	//		log.Printf("failed to recv: %v", streamError)
	//	}
	//	fmt.Printf("StringService Concat : %s concat %s = %s\n", stringReq.A, stringReq.B, item.GetRet())
	//}

	//3.流请求:客户端流式请求，服务端整合流式请求后，返回一个响应结果（字符串）
	//sendClientStreamRequest(stringClient)

	//4.双向流请求：客户端发送一个请求后，立刻就可以获取对应服务端的响应
	sendClientAndServerStreamRequest(stringClient)
}

// 客户端流和双向流的区别： 客户端流rpc回先将客户端的请求以流的形式发送完毕后，再获取服务端的响应流
// 而双向流rpc，客户端发送一个请求后，立刻就可以获取对应服务端的响应

//客户端流式请求，服务端整合流式请求后，返回一个响应结果（字符串）
func sendClientStreamRequest(client stream_pb.StringServiceClient) {
	fmt.Printf("test sendClientStreamRequest \n")

	stream, err := client.LotsOfClientStream(context.Background())
	//发起流rpc请求后，获得一个stream结构体
	for i := 0; i < 20; i++ {
		if err != nil {
			log.Printf("failed to call: %v", err)
			break
		}
		stream.Send(&stream_pb.StringRequest{A: strconv.Itoa(i), B: strconv.Itoa(i) })
	}
	// 发送完成后，可以调用CloseAndRecv方法来获取服务端的请求响应
	reply, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("failed to recv: %v", err)
	}
	log.Printf("sendClientStreamRequest ret is : %s", reply.Ret)
}

//双向流：客户端和服务端均在同一个stream流内进行数据交互（发送两个字符串，返回拼接结果）
func sendClientAndServerStreamRequest(client stream_pb.StringServiceClient) {
	fmt.Printf("test sendClientAndServerStreamRequest \n")
	var err error
	stream, err := client.LotsOfServerAndClientStream(context.Background())
	if err != nil {
		log.Printf("failed to call: %v", err)
		return
	}
	var i int
	for {
		fmt.Println("id: ",i)
		err1 := stream.Send(&stream_pb.StringRequest{A: strconv.Itoa(i), B: strconv.Itoa(i + 1)})
		if err1 != nil {
			log.Printf("failed to send: %v", err)
			break
		}
		reply, err2 := stream.Recv()
		if err2 != nil {
			log.Printf("failed to recv: %v", err)
			break
		}
		log.Printf("sendClientAndServerStreamRequest Ret is : %s", reply.Ret)
		i++
	}
}
