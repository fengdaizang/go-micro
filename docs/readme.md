## go-micro项目大纲

### 1.config配置文件

#### 1.1.配置文件名称

默认为config，可在common/starter/starter.go中修改常量`configFileName`的值即可。

#### 1.2.配置文件位置

- 如果在初始化配置时指定了配置文件地址，则会使用该值。在common/starter/starter.go中initConfig方法处，会指定初始化配置文件的配置。

- 其次会读取环境变量`${constant.ServiceEnvVarPrefix}${serviceName}_CFG_PATH`的值。
  - 一般会在使用镜像运行项目时用到，此时会在Dockerfile中定义该环境变量的值，如`ENV SPDB_EVIDENCE_CFG_PATH /data/spdb/private-placement/evidence/config/`
  - `${constant.ServiceEnvVarPrefix}`的值为common/constant/constant.go中定义的服务环境变量前缀。
  - `${serviceName}`的值为main.go中调用`starter.StartAPIServer`传递的服务名。
- 如果以上步骤均为取到值，则会直接读取项目的config目录。

#### 1.3.多环境配置切换

可以通过在配置文件config.yaml中指定启用某几个环境的配置文件进行环境切换。配置项为`service.profiles.active`。环境对应的配置文件命名规则为`${configName}_${env}`，`${configName}`为原配置文件名称，`${env}`为环境名称，如local、uat、rel等。

