local http = require("http");
local json = require("json");

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
    return 'test', userid
end

-- 接收提现请求
-- @param to <string> 目标地址
-- @param symbol <string> 货币符号
-- @param amount <string> 提现金额
-- @return txid <string> 交易ID
-- @return error <string or nil> 错误信息
function on_withdraw(to, symbol, amount)
    future:set_result(nil, 'unrealized')
end

-- 交易是否有效
-- @param txid <string> 交易ID
-- @param from <string> 来源地址
-- @param to <string> 目标地址
-- @param symbol <string> 货币符号
-- @param amount <string> 交易金额
-- @param memo <string> 备注信息
-- @return <boolean>
function valid_transaction(txid, from, to, symbol, amount, memo)
    return true
end
