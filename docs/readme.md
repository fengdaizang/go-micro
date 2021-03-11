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

### 2.log日志

可以通过配置文件进行相关配置，目前支持配置日志级别、日志时间格式、是否输出到文件等配置。通过`logging.file.enable`来开启或关闭输出到文件。

如果开启输出至文件，需要配置文件输出路径，不配置则默认为项目路径。支持文件分割，相关配置在common/log/log.go的initLoggerParams方法中默认指定的。文件名默认为`${serviceName}.${yyyyMMdd}.log`，分割时间为一天一个文件，日志保留时间为一周。

注意事项：如果日志分割时间内，日志文件名称一致的话，则不会产生新文件，而是会追加在后面，因此需要注意文件命名规则是否合适，文件名变化周期应小于等于分割时间。

也可以采用设置文件大小的方式进行分割，这样分割后文件名默认为`${serviceName}.${yyyyMMdd}.log.${number}`

### 3.Starter启动器

通过调用` starter.StartAPIServer(serviceName string,version. string)`方法获取一个webService，传递参数为服务名称、版本号，该方法会读取配置文件、配置日志、如果开启注册中心则会像注册中心进行服务发现与注册。

### 4.Service服务

会将API接口进行初始化，包括初始化API接口以及中间件处理器等。

