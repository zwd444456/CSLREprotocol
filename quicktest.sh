
# 设置Go环境变量（如果需要）
# export GOROOT=/usr/local/go
# export GOPATH=$HOME/go
# export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# 导航到项目根目录
cd "$(dirname "$0")"

# 打印当前工作目录
echo "当前工作目录: $(pwd)"

# 编译Go文件
echo "正在编译 nomal2PC/nomal2pc.go ..."
go build -o nomal2pc nomal2PC/nomal2pc.go

# 检查编译是否成功
if [ $? -ne 0 ]; then
  echo "编译失败，请检查Go代码。"
  exit 1
fi

# 运行编译后的二进制文件
echo "正在运行 nomal2pc ..."
./nomal2pc/nomal2pc

# 检查运行是否成功
if [ $? -ne 0 ]; then
  echo "程序运行失败。"
  exit 1
fi

# 暂停以便查看输出
read -p "按任意键退出..."
