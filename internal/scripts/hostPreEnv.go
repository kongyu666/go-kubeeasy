package scripts

import "github.com/minph/fs"

const contentHostPreEnv = `#!/usr/bin/env bash
# 节点初始化脚本
# Disable selinux
sed -i 's/SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
setenforce 0

# Disable swap
swapoff -a
sed -ri '/^[^#]*swap/s@^@#@' /etc/fstab

# Disable firewalld
for target in firewalld python-firewall firewalld-filesystem iptables; do
	systemctl stop $target &>/dev/null || true
	systemctl disable $target &>/dev/null || true
done

# ssh
# 关闭反向解析，加快连接速度
sed -i \
	-e 's/#UseDNS yes/UseDNS no/g' \
	-e 's/GSSAPIAuthentication yes/GSSAPIAuthentication no/g' \
	/etc/ssh/sshd_config
# 取消确认键
sed -i 's/#   StrictHostKeyChecking ask/   StrictHostKeyChecking no/g' /etc/ssh/ssh_config
systemctl restart sshd
  
  # Change limits
  [ ! -f /etc/security/limits.conf_bak ] && cp /etc/security/limits.conf{,_bak}
  cat << EOF >> /etc/security/limits.conf
## kubeeasy managed start
root soft nofile 655360
root hard nofile 655360
root soft nproc 655360
root hard nproc 655360
root soft core unlimited
root hard core unlimited
* soft nofile 655360
* hard nofile 655360
* soft nproc 655360
* hard nproc 655360
* soft core unlimited
* hard core unlimited
## kubeeasy managed end
EOF

  [ -f /etc/security/limits.d/20-nproc.conf ] && sed -i 's#4096#655360#g' /etc/security/limits.d/20-nproc.conf
  cat << EOF >> /etc/systemd/system.conf
## kubeeasy managed start
DefaultLimitCORE=infinity
DefaultLimitNOFILE=655360
DefaultLimitNPROC=655360
DefaultTasksMax=75%
## kubeeasy managed end
EOF

   # Change sysctl
   cat << EOF >  /etc/sysctl.d/99-kubeeasy.conf
# https://www.kernel.org/doc/Documentation/sysctl/
# 开启IP转发.
net.ipv4.ip_forward = 1
# 要求iptables不对bridge的数据进行处理
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-arptables = 1
# vm.max_map_count 计算当前的内存映射文件数。
# mmap 限制（vm.max_map_count）的最小值是打开文件的ulimit数量（cat /proc/sys/fs/file-max）。
# 每128KB系统内存 map_count应该大约为1。 因此，在32GB系统上，max_map_count为262144。
# Default: 65530
vm.max_map_count = 262144
# Default: 30
# 0 - 任何情况下都不使用swap。
# 1 - 除非内存不足（OOM），否则不使用swap。
vm.swappiness = 0
# 文件监控
fs.inotify.max_user_instances=524288
fs.inotify.max_user_watches=524288
fs.inotify.max_queued_events=16384
# 调高 PID 数量
kernel.pid_max = 65536
kernel.threads-max=30938
EOF

  # history
  cat <<EOF >>/etc/bashrc
## kubeeasy managed start
# history actions record，include action time, user, login ip
HISTFILESIZE=100000
HISTSIZE=100000
USER_IP=\$(who -u am i 2>/dev/null | awk '{print \$NF}' | sed -e 's/[()]//g')
if [ -z \$USER_IP ]
then
  USER_IP=\$(hostname -i)
fi
HISTTIMEFORMAT="%Y-%m-%d %H:%M:%S \$USER_IP:\$(whoami) "
HISTFILE=~/.bash_history
shopt -s histappend
PROMPT_COMMAND="history -a"
export HISTFILESIZE HISTSIZE HISTTIMEFORMAT HISTFILE PROMPT_COMMAND

# PS1
PS1='\[\033[0m\]\[\033[1;36m\][\u\[\033[0m\]@\[\033[1;32m\]\h\[\033[0m\] \[\033[1;31m\]\W\[\033[0m\]\[\033[1;36m\]]\[\033[33;1m\]\\$ \[\033[0m\]'
## kubeeasy managed end
EOF


   # journal
   mkdir -p /var/log/journal /etc/systemd/journald.conf.d
   cat << EOF > /etc/systemd/journald.conf.d/99-prophet.conf
[Journal]
# 持久化保存到磁盘
Storage=persistent
# 压缩历史日志
Compress=yes
SyncIntervalSec=5m
RateLimitInterval=30s
RateLimitBurst=1000
# 最大占用空间 10G
SystemMaxUse=10G
# 单日志文件最大 200M
SystemMaxFileSize=200M
# 日志保存时间 3 周
MaxRetentionSec=3week
# 不将日志转发到 syslog
ForwardToSyslog=no
EOF

  # motd
  cat <<EOF >/etc/profile.d/ssh-login-info.sh
#!/bin/sh
#
# @Time    : 2022-04-13
# @Author  : KongYu
# @Desc    : ssh login banner

export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
shopt -q login_shell && : || return 0
echo -e "\033[1;3\$((RANDOM%10%8))m

  ██╗  ██╗ █████╗ ███████╗
  ██║ ██╔╝██╔══██╗██╔════╝
  █████╔╝ ╚█████╔╝███████╗
  ██╔═██╗ ██╔══██╗╚════██║
  ██║  ██╗╚█████╔╝███████║
  ╚═╝  ╚═╝ ╚════╝ ╚══════╝ \033[0m"


# os
upSeconds="\$(cut -d. -f1 /proc/uptime)"
secs=\$((\${upSeconds}%60))
mins=\$((\${upSeconds}/60%60))
hours=\$((\${upSeconds}/3600%24))
days=\$((\${upSeconds}/86400))
UPTIME_INFO=\$(printf "%d days, %02dh %02dm %02ds" "\$days" "\$hours" "\$mins" "\$secs")

if [ -f /etc/redhat-release ] ; then
    PRETTY_NAME=\$(< /etc/redhat-release)

elif [ -f /etc/debian_version ]; then
   DIST_VER=\$(</etc/debian_version)
   PRETTY_NAME="\$(grep PRETTY_NAME /etc/os-release | sed -e 's/PRETTY_NAME=//g' -e  's/"//g') (\$DIST_VER)"

else
    PRETTY_NAME=\$(cat /etc/*-release | grep "PRETTY_NAME" | sed -e 's/PRETTY_NAME=//g' -e 's/"//g')
fi

if [[ -d "/system/app/" && -d "/system/priv-app" ]]; then
    model="\$(getprop ro.product.brand) \$(getprop ro.product.model)"

elif [[ -f /sys/devices/virtual/dmi/id/product_name ||
        -f /sys/devices/virtual/dmi/id/product_version ]]; then
    model="\$(< /sys/devices/virtual/dmi/id/product_name)"
    model+="\$(< /sys/devices/virtual/dmi/id/product_version)"

elif [[ -f /sys/firmware/devicetree/base/model ]]; then
    model="\$(< /sys/firmware/devicetree/base/model)"

elif [[ -f /tmp/sysinfo/model ]]; then
    model="\$(< /tmp/sysinfo/model)"
fi

MODEL_INFO=\${model}
KERNEL=\$(uname -srmo)
USER_NUM=\$(who -u | wc -l)
RUNNING=\$(ps ax | wc -l | tr -d " ")

# disk total
totaldisk=\$(df -h -x devtmpfs -x tmpfs -x debugfs -x aufs -x overlay --total 2>/dev/null | tail -1)
disktotal=\$(awk '{print \$2}' <<< "\${totaldisk}")
diskused=\$(awk '{print \$3}' <<< "\${totaldisk}")
diskusedper=\$(awk '{print \$5}' <<< "\${totaldisk}")
DISK_INFO="\033[0;33m\${diskused}\033[0m/\033[1;34m\${disktotal}\033[0m (\033[0;33m\${diskusedper}\033[0m)"

# disk root
totaldisk_root=\$(df -h -x devtmpfs -x tmpfs -x debugfs -x aufs -x overlay --total 2>/dev/null | egrep -v Filesystem | head -n1)
disktotal_root=\$(awk '{print \$2}' <<< "\${totaldisk_root}")
diskused_root=\$(awk '{print \$3}' <<< "\${totaldisk_root}")
diskusedper_root=\$(awk '{print \$5}' <<< "\${totaldisk_root}")
DISK_INFO_ROOT="\033[0;33m\${diskused_root}\033[0m/\033[1;34m\${disktotal_root}\033[0m (\033[0;33m\${diskusedper_root}\033[0m)"

# cpu
cpu=\$(awk -F':' '/^model name/ {print \$2}' /proc/cpuinfo | uniq | sed -e 's/^[ \t]*//')
cpun=\$(grep -c '^processor' /proc/cpuinfo)
cpuc=\$(grep '^cpu cores' /proc/cpuinfo | tail -1 | awk '{print \$4}')
cpup=\$(grep '^physical id' /proc/cpuinfo | wc -l)
MODEL_NAME=\$cpu
CPU_INFO="\$(( cpun*cpuc ))(cores) \$(grep '^cpu MHz' /proc/cpuinfo | tail -1 | awk '{print \$4}')(MHz) \${cpup}P(physical) \${cpuc}C(cores) \${cpun}L(processor)"

# cpu usage
CPU_USAGE=\$(echo 100 - \$(top -b -n 1 | grep Cpu | awk '{print \$8}') | bc)

# get the load averages
read one five fifteen rest < /proc/loadavg
LOADAVG_INFO="\033[0;33m\${one}\033[0m(1min) \033[0;33m\${five}\033[0m(5min) \033[0;33m\${fifteen}\033[0m(15min)"

# mem
MEM_INFO="\$(cat /proc/meminfo | awk '/MemTotal:/{total=\$2/1024/1024;next} /MemAvailable:/{use=total-\$2/1024/1024; printf("\033[0;33m%.2fGiB\033[0m/\033[1;34m%.2fGiB\033[0m (\033[0;33m%.2f%%\033[0m)",use,total,(use/total)*100);}')"

# network
# extranet_ip=" and \$(curl -s ip.cip.cc)"
IP_INFO="\$(ip a | grep glo | awk '{print \$2}' | head -1 | cut -f1 -d/)\${extranet_ip:-}"

# Container info
CONTAINER_INFO=\$(sudo /usr/bin/crictl ps -a -o yaml 2> /dev/null | awk '/^  state: /{gsub("CONTAINER_", "", \$NF) ++S[\$NF]}END{for(m in S) printf "%s%s:%s ",substr(m,1,1),tolower(substr(m,2)),S[m]}')Images:\$(sudo /usr/bin/crictl images -q 2> /dev/null | wc -l)

# info
echo -e "
 Information as of: \033[1;34m\$(date +"%Y-%m-%d %T")\033[0m

 \033[0;1;31mProduct\033[0m............: \${MODEL_INFO}
 \033[0;1;31mOS\033[0m.................: \${PRETTY_NAME}
 \033[0;1;31mKernel\033[0m.............: \${KERNEL}
 \033[0;1;31mCPU Model Name\033[0m.....: \${MODEL_NAME}
 \033[0;1;31mCPU Cores\033[0m..........: \${CPU_INFO}

 \033[0;1;31mHostname\033[0m...........: \033[1;34m\$(hostname)\033[0m
 \033[0;1;31mIP Addresses\033[0m.......: \033[1;34m\${IP_INFO}\033[0m

 \033[0;1;31mUptime\033[0m.............: \033[0;33m\${UPTIME_INFO}\033[0m
 \033[0;1;31mMemory Usage\033[0m.......: \${MEM_INFO}
 \033[0;1;31mCPU Usage\033[0m..........: \033[0;33m\${CPU_USAGE}%\033[0m
 \033[0;1;31mLoad Averages\033[0m......: \${LOADAVG_INFO}
 \033[0;1;31mDisk Total Usage\033[0m...: \${DISK_INFO}
 \033[0;1;31mDisk Root Usage\033[0m....: \${DISK_INFO_ROOT}

 \033[0;1;31mUsers online\033[0m.......: \033[1;34m\${USER_NUM}\033[0m
 \033[0;1;31mRunning Processes\033[0m..: \033[1;34m\${RUNNING}\033[0m
 \033[0;1;31mContainer Info\033[0m.....: \${CONTAINER_INFO}
"
EOF

chmod +x /etc/profile.d/ssh-login-info.sh
echo 'ALL ALL=(ALL) NOPASSWD:/usr/bin/crictl' >/etc/sudoers.d/crictl

# sync time
timedatectl set-timezone Asia/Shanghai
timedatectl set-local-rtc 1
hwclock --systohc
date
hwclock -r

sysctl --system
`

func HostPreEnv(scriptFile string) {
	fs.Create(scriptFile)
	fs.Rewrite(scriptFile, contentHostPreEnv)
}
