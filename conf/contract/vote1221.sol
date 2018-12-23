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
     * @param option_ids 字符串数组类型数据
     * @param option_contents 字符串数组整类型数据
     *
     * @return int32 返回代码
     * @return bytes 返回消息
     */
    function insertVote(bytes32 id, bytes32 title, bytes32 description, int32 select_type, bytes32 start_time, bytes32 end_time
    , bytes32 create_time, bytes32 creator_id, bytes32[] option_ids, bytes32[] option_contents) public returns(int32, bytes) {

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
            return (ERROR, "主键已经存在，无法插入");
        }
        // 存储主键
        _idInVoteArray.push(newVote.id);
        //存储数据
        _id2Vote[newVote.id] = newVote;
        // 累计插入数量
        insertedCount = insertedCount + 1;
        
        // 按主键插入多条投票选项内容
        if(option_ids.length != 0){
            uint length = option_ids.length;
            for(uint i = 0; i < length; i++) {
                insertVoteOption(option_ids[i], newVote.id, option_contents[i]);
            }
        }

        return (SUCCESS, "插入成功");
    }

    /**
     * @dev 按主键查询多条投票内容表
     *
     * @param id 字符串类型数据
     *
     * @return int32 返回代码
     * @return bytes 返回标题
     * @return bytes 返回描述
     * @return int32 返回单选/多选
     * @return bytes 返回开始时间
     * @return bytes 返回结束时间
     * @return bytes 返回创建时间
     * @return bytes 返回创建者ID
     */
    function queryVote(bytes32 id) public returns(int32, bytes32 title, bytes32 description,
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
            return (SUCCESS, title, description, select_type, start_time, end_time, create_time, creator_id);
        }
        return (ERROR, title, description, select_type, start_time, end_time, create_time, creator_id);
    }


/***********************************************************************************************************************
                                                      投票选项内容
 **********************************************************************************************************************/
    struct VoteOption {
    bytes32 id;           //主键
    bytes32 vote_id;      //所属投票的ID
    bytes32 content;      //内容
    int32  total;          //票数
    }

    // 主键2结构体
    mapping (bytes32 => VoteOption) _id2VoteOption;

    // 所有主键
    bytes32[] _idInVoteOptionArray;

    //投票选项所属的投票活动ID
    mapping (bytes32 => bytes32[]) _optionID2Vote;

    /**
     * @dev 按主键插入多条投票选项内容
     *
     * @param id 字符串类型数据
     * @param vote_id 字符串类型数据
     * @param content 字符串类型数据
     *
     * @return int32 返回代码
     * @return bytes 返回消息
     */
    function insertVoteOption(bytes32 id, bytes32 vote_id, bytes32 content) public returns(int32, bytes) {

        VoteOption memory newVoteOption;
        // 从入参中解析出数据
        newVoteOption.id = id;
        newVoteOption.vote_id = vote_id;
        newVoteOption.content = content;
        newVoteOption.total = 0;
        // 若主键存在则不插入
        if (_id2VoteOption[newVoteOption.id].id != 0) {
            return (ERROR, "主键已经存在，无法插入");
        }
        //若复合主键voteID2VoteDetail存在则不插入
        if(_id2Vote[newVoteOption.vote_id].id == 0){
            return (ERROR, "复合主键voteID2VoteDetail不存在，无法插入");
        }
        // 存储主键
        _idInVoteOptionArray.push(newVoteOption.id);
        //存储数据
        _id2VoteOption[newVoteOption.id] = newVoteOption;
        // 存储选项ID到对应的vote数组
        _optionID2Vote[newVoteOption.vote_id].push(newVoteOption.id);

        return (SUCCESS, "插入成功");
    }

    /**
     * @dev 按主键更新多条投票选项内容
     *
     * @param id 字符串类型数据
     *
     * @return int32 返回代码
     * @return bytes 返回消息
     */
    function updateVoteOption(bytes32 id) public returns(int32, bytes) {

        VoteOption memory newVoteOption;
        VoteOption memory oldVoteOption;
        // 从入参中解析出数据
        newVoteOption.id = id;
        // 若主键不存在则不更新
        oldVoteOption.id = _id2VoteOption[newVoteOption.id].id;
        if (oldVoteOption.id == 0) {
            return (ERROR, "主键不存在，无法更新");
        }
        //从原有数据中取出部分用于更新
        oldVoteOption.total = _id2VoteOption[newVoteOption.id].total;
        newVoteOption.total = oldVoteOption.total + 1;
        newVoteOption.vote_id = _id2VoteOption[newVoteOption.id].vote_id;
        newVoteOption.content = _id2VoteOption[newVoteOption.id].content;
        // 存储数据
        _id2VoteOption[newVoteOption.id] = newVoteOption;
        return (SUCCESS, "更新成功");
    }

    /**
     * @dev 按主键查询多条投票内容表
     *
     * @param id 字符串类型数据
     *
     * @return int32 返回代码
     * @return bytes32[] 返回选项ID
     * @return bytes32[] 返回选项内容数组
     * @return int32[] 返回投票结果内容数组
     */
    function queryVoteOption(bytes32 id) public returns(int32, bytes32[], bytes32[] , int32[] ) {

        VoteOption memory voteOption;

        initArrayReturn();

        // 从入参中解析出数据
        if(_optionID2Vote[id].length != 0){
            uint length = _optionID2Vote[id].length;
            bytes32[] optionIds  = _optionID2Vote[id];
            for(uint i = 0; i < length; i++) {
                bytes32 option_id = optionIds[i];
                voteOption = _id2VoteOption[option_id];
                _bytes32ArrayReturn.push(voteOption.content);
                _intArrayReturn.push(voteOption.total);
            }
            return(SUCCESS, optionIds, _bytes32ArrayReturn, _intArrayReturn);
        }
        return (ERROR, _bytes32ArrayReturn, _bytes32ArrayReturn, _intArrayReturn);
    }

/***********************************************************************************************************************
                                                        投票记录
 **********************************************************************************************************************/
    struct VoteResult {
    bytes32 id;             //主键
    bytes32 vote_id;        //投票活动ID
    bytes32 option_id;      //投票选项ID
    bytes32 option_content; //选项内容
    bytes32 user_id;        //用户ID
    bytes32 public_key;     //用户公钥
    bytes32 create_time;    //投票时间
    }

    // 主键2结构体
    mapping (bytes32 => VoteResult) _id2VoteResult;

    // 所有主键
    bytes32[] _idInVoteResultArray;

     // 主键2结构体
    mapping (bytes32 => bytes32[]) _userIds2VoteResult;

    /**
     * @dev 按主键插入多条投票记录
     *
     * @param id 字符串类型数据
     * @param vote_id 字符串类型数据
     * @param option_id 字符串类型数据
     * @param option_content 字符串类型数据
     * @param user_id 字符串类型数据
     * @param public_key 字符串类型数据
     * @param create_time 字符串类型数据
     *
     * @return int32 返回代码
     * @return bytes 返回消息
     */
    function insertVoteResult(bytes32 id, bytes32 vote_id, bytes32 option_id, bytes32 option_content, bytes32 user_id,
        bytes32 public_key, bytes32 create_time) public returns(int32, bytes) {

        VoteResult memory newVoteResult;

        // 从入参中解析出数据
        newVoteResult.id = id;
        newVoteResult.vote_id = vote_id;
        newVoteResult.option_id = option_id;
        newVoteResult.option_content = option_content;
        newVoteResult.user_id = user_id;
        newVoteResult.public_key = public_key;
        newVoteResult.create_time = create_time;
        // 若主键存在则不插入
        if (_id2VoteResult[newVoteResult.id].id != 0) {
            return (ERROR, "主键已经存在，无法插入");
        }
        //若复合主键voteID2VoteResult存在则不插入
        if(_id2VoteOption[newVoteResult.option_id].id == 0){
            return (ERROR, "复合主键option_id不存在，无法插入");
        }

        // 存储主键
        _idInVoteResultArray.push(newVoteResult.id);
        //存储数据
        _id2VoteResult[newVoteResult.id] = newVoteResult;
        // 存储用户的投票
        _userIds2VoteResult[newVoteResult.user_id].push(newVoteResult.id);

        return (SUCCESS, "插入成功");
    }

        /**
     * @dev 按主键查询多条投票内容表
     *
     * @param user_id 字符串类型数据
     * @param vote_id 字符串类型数据
     *
     * @return int32 返回代码
     * @return bool  返回投票结果
     */
    function queryVoteResult(bytes32 user_id, bytes32 vote_id) public returns(int32, bool) {

        VoteResult memory voteResult;


        // 从入参中解析出数据
        if(_userIds2VoteResult[user_id].length != 0){
            uint length = _userIds2VoteResult[user_id].length;
            bytes32[] voteResultIds  = _userIds2VoteResult[user_id];
            for(uint i = 0; i < length; i++) {
                bytes32 voteResultId = voteResultIds[i];
                voteResult = _id2VoteResult[voteResultId];
                if(voteResult.vote_id == vote_id){
                    return (SUCCESS, true);
                }
            }
            return(SUCCESS, false);
        }
        return (ERROR, false);
    }

/***********************************************************************************************************************
                                                        全局常量
 **********************************************************************************************************************/

    // 返回代码常量：成功（0）
    int32 constant SUCCESS = 0;

    // 返回代码常量：业务逻辑错误（1）
    int32 constant ERROR = 1;

/***********************************************************************************************************************
                                                        内部方法
 **********************************************************************************************************************/

    bytes32[] _bytes32ArrayReturn;

    uint[] _uintArrayReturn;

    int32[] _intArrayReturn;

    address[] _addressArrayReturn;

    function initArrayReturn() internal {
        _bytes32ArrayReturn.length = 0;
        _uintArrayReturn.length = 0;
        _intArrayReturn.length = 0;
        _addressArrayReturn.length = 0;
    }

    function bytes32ArrayReturnPush(bytes32[] storage array) internal {
        uint length = array.length;
        for (uint i = 0; i < length; i = i + 1) {
            _bytes32ArrayReturn.push(array[i]);
        }
        _uintArrayReturn.push(length);
    }

    function uintArrayReturnPush(uint[] storage array) internal {
        uint length = array.length;
        for (uint i = 0; i < length; i = i + 1) {
            _uintArrayReturn.push(array[i]);
        }
        _uintArrayReturn.push(length);
    }

    function intArrayReturnPush(int32[] storage array) internal {
        uint length = array.length;
        for (uint i = 0; i < length; i = i + 1) {
            _intArrayReturn.push(array[i]);
        }
        _uintArrayReturn.push(length);
    }

    function addressArrayReturnPush(address[] storage array) internal {
        uint length = array.length;
        for (uint i = 0; i < length; i = i + 1) {
            _addressArrayReturn.push(array[i]);
        }
        _uintArrayReturn.push(length);
    }

}