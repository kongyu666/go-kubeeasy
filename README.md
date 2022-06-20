
## 安装go-kubeeasy
```shell
curl -LO https://github.com/kongyu666/go-kubeeasy/releases/download/v1.0.0/go-kubeeasy
chmod +x go-kubeeasy && mv go-kubeeasy /usr/loca/bin/
```

## 设置CentOS系统基础环境
```shell
go-kubeeasy install hostpre 192.168.1.202 192.168.1.202 -u root -p Admin@123
```

## 执行指定的shell命令
```shell
go-kubeeasy cmd 192.168.1.201 192.168.1.202 -e w -u root -p Admin@123
```

## 执行指定的shell脚本
```shell
go-kubeeasy install script 192.168.1.201 192.168.1.202 -e test.sh -u root -p Admin@123
```
