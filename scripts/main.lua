local http = require("http")

-- 时钟循环
function on_tick(delaytime)
end

-- 账户是否有效
function valid_account(account)
    return true
end

-- 接收提现请求
function on_withdraw(to, symbol, amount, future)
    print('on_withdraw', to, symbol, amount, future)
    future:set_result(nil)
    -- future:set_result('failed to transfer')
end

-- 交易是否有效
function valid_transaction(from, to, symbol, amount, txid)
    return true
end
