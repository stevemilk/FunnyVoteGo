pragma solidity ^0.4.10;

/**
 * VoteContract项目智能合约源码：投票合约
 *
 * Copyright(C)2016-2018 Hyperchain Technologies Co.,Ltd. All rights reserved.
 *
 * 2018-12-18 13:12:15
 */
contract VoteContract {


/***********************************************************************************************************************
                                                       投票内容表
 **********************************************************************************************************************/
    struct Vote {
    bytes32 id;              //主键
    bytes32 title;           //投票名字
    bytes32 description;     //投票描述
    int32 select_type;     //单选/多选
    bytes32 start_time;      //开始时间
    bytes32 end_time;        //结束时间
    bytes32 create_time;     //创建时间
    bytes32 creator_id;      //创建者ID
    }

    // 主键2结构体
    mapping (bytes32 => Vote) _id2Vote;

    // 所有主键
    bytes32[] _idInVoteArray;
    /**
     * @dev 按主键插入多条投票内容表
     *
     * @param id 字符串类型数据
     * @param title 字符串类型数据
     * @param description 字符串类型数据
     * @param select_type 整数类型数据
     * @param start_time 字符串类型数据
     * @param end_time 字符串类型数据
     * @param create_time 字符串类型数据
     * @param creator_id 整数类型数据
     *
     * @return uint 返回代码
     * @return bytes 返回消息
     * @return uint 本次插入成功的数据量
     */
    function insertVote(bytes32 id, bytes32 title, bytes32 description, int32 select_type, bytes32 start_time, bytes32 end_time
    , bytes32 create_time, bytes32 creator_id) public returns(bytes) {

        Vote memory newVote;
        uint insertedCount = 0;
        // 从入参中解析出数据
        newVote.id = id;
        newVote.title = title;
        newVote.description = description;
        newVote.select_type = select_type;
        newVote.start_time = start_time;
        newVote.end_time = end_time;
        newVote.create_time = create_time;
        newVote.creator_id = creator_id;
        // 若主键存在则不插入
        if (_id2Vote[newVote.id].id != 0) {
            return ("主键已经存在，无法插入");
        }
        // 存储主键
        _idInVoteArray.push(newVote.id);
        //存储数据
        _id2Vote[newVote.id] = newVote;
        // 累计插入数量
        insertedCount = insertedCount + 1;

        return ("插入成功");
    }

    /**
     * @dev 按主键查询多条投票内容表
     *
     * @param id 字符串类型数据
     *
     * @return uint 返回代码
     * @return bytes 返回消息
     * @return bytes32[] 字符串类型数据
     * @return int[] 整数类型数据
     */
    function queryVote(bytes32 id) public returns(bool, bytes32 title, bytes32 description,
    int32 select_type, bytes32 start_time, bytes32 end_time, bytes32 create_time,
    bytes32 creator_id) {

        Vote memory oldVote;

        // 从入参中解析出数据

        oldVote.id = id;
        if (oldVote.id != 0) {
            title = _id2Vote[oldVote.id].title;
            description = _id2Vote[oldVote.id].description;
            select_type = _id2Vote[oldVote.id].select_type;
            start_time = _id2Vote[oldVote.id].start_time;
            end_time = _id2Vote[oldVote.id].end_time;
            create_time = _id2Vote[oldVote.id].create_time;
            creator_id = _id2Vote[oldVote.id].creator_id;
            return (true, title, description, select_type, start_time, end_time, create_time, creator_id);
        }
        return (false, title, description, select_type, start_time, end_time, create_time, creator_id);
    }


/***********************************************************************************************************************
                                                      多表查询汇总
 **********************************************************************************************************************/


/***********************************************************************************************************************
                                                        全局常量
 **********************************************************************************************************************/

    // 返回代码常量：成功（0）
    uint constant SUCCESS = 0;

    // 返回代码常量：业务逻辑错误（1）
    uint constant ERROR = 1;

}