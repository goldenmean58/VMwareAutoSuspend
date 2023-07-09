在关机或重启时自动suspend所有在运行的VMware虚拟机

使用：
1. 将编译出的exe文件放到需要的目录
2. 开启管理员权限，用 `xxx.exe install` 的方式安装服务
3. 打开 `services.msc` 找到 `VMwareAutoSuspender` 服务，设置它使用的账户为当前运行虚拟机的账户
4. 启动服务

不会和VmwareAutostartService冲突，可以同时使用

----

原理：

利用 [SERVICE_ACCEPT_PRESHUTDOWN](https://www.coretechnologies.com/blog/windows-services/increase-shutdown-time/) 机制，通过注册一个服务，在关机时执行[相关命令](https://github.com/fatso83/vmware-auto-suspend/blob/master/SuspendRunningVMs.bat)来暂停