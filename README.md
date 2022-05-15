## Client-go
```text
1. getkube 获取kubeconfig配置文件信息
2. restclient 通过restclient获取kube-system下面的pod列表
3. clientCreateDeploy 通过cilentset创建deploy
4. shareProcessNs 
   在进行pod故障诊断时 往往会发现 基础镜像缺少部分命令，通过在deploy的annotations 中设置[shell:"true"]参数，共享命名空间 来实现在原有deploy中增加一个shell的container，
   通过kubectl exec deploy -c shell -- sh进入容器进行故障拍错
```


## kubebuilder
```text
1. b kubebuilder example
```
