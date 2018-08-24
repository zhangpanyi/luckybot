-- 时钟事件
-- @param delaytime <number>
function on_tick(delaytime)
end

-- 账户是否有效
-- @param address <string> 地址
-- @return <boolean>
function valid_address(address)
    return true
end

-- 获取充值地址
-- @param userid <string> 用户ID
-- @return address <string>
-- @return memo <string or nil>
function deposit_address(userid)
    return 'bts', nil
end

-- 接收提现请求
-- @param to <string> 目标地址
-- @param symbol <string> 货币符号
-- @param amount <string> 提现金额
-- @param future <Future> 异步任务
function on_withdraw(to, symbol, amount, future)
    future:set_result(nil)
end

-- 交易是否有效
-- @param from <string> 来源地址
-- @param to <string> 目标地址
-- @param symbol <string> 货币符号
-- @param amount <string> 交易金额
-- @param memo <string> 备注信息
-- @param txid <string> 交易ID
-- @return <boolean>
function valid_transaction(from, to, symbol, amount, memo, txid)
    return true
end
