package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/OPC-16/go-orders-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
    Client *redis.Client
}

//this func takes id and outputs formatted string in form `order:{id}`,
//we need this coz redis stores data in key-value pair format
func orderIDKey(id uint64) string {
    return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
    data, err := json.Marshal(order)
    if err != nil {
        return fmt.Errorf("failed to encode order: %w", err)
    }

    key := orderIDKey(order.OrderID)

    //this starts our new transaction, we want to do 2 things first store data in DB and second add same data to 'orders' set,
    //thats why we need to do it in a transaction so as not to end up in a partial state
    txn := r.Client.TxPipeline()

    //SetNX methods inserts data only if it previously didn't exist in DB
    res := txn.SetNX(ctx, key, string(data), 0)
    if err := res.Err(); err != nil {
        txn.Discard()
        return fmt.Errorf("failed to set: %w", err)
    }

    if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
        txn.Discard()
        return fmt.Errorf("failed to add to orders set: %w", err)
    }

    //Exec() commits our changes to the DB
    if _, err := txn.Exec(ctx); err != nil {
        return fmt.Errorf("failed to Exec or Commit: %w", err)
    }

    return nil
}

//we created our custom error for any order that doesn't exist
var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
    key := orderIDKey(id)

    value, err := r.Client.Get(ctx, key).Result()
    if errors.Is(err, redis.Nil) {
        return model.Order{}, ErrNotExist
    } else if err != nil {
        return model.Order{}, fmt.Errorf("get order: %w", err)
    }

    var order model.Order
    err = json.Unmarshal([]byte(value), &order)     //this unmarshals or decodes the json data which is stored in value and puts it in order
    if err != nil {
        return model.Order{}, fmt.Errorf("failed to decode the json: %w", err)
    }

    return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
    key := orderIDKey(id)

    txn := r.Client.TxPipeline()

    err := txn.Del(ctx, key).Err()
    if errors.Is(err, redis.Nil) {
        txn.Discard()
        return ErrNotExist
    } else if err != nil {
        txn.Discard()
        return fmt.Errorf("get order: %w", err)
    }

    if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
        txn.Discard()
        return fmt.Errorf("failed to remove from orders set: %w", err)
    }

    if _, err := txn.Exec(ctx); err != nil {
        return fmt.Errorf("failed to exec: %w", err)
    }

    return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
    data, err := json.Marshal(order)
    if err != nil {
        return fmt.Errorf("failed to encode order: %w", err)
    }

    key := orderIDKey(order.OrderID)

    //SetXX method only updates any existing data
    err = r.Client.SetXX(ctx, key, string(data), 0).Err()
    if errors.Is(err, redis.Nil) {
        return ErrNotExist
    } else if err != nil {
        return fmt.Errorf("set order: %w", err)
    }

    return nil
}

//to support pagination
type FindAllPage struct {
    Size    uint64
    Offset  uint64
}

type FindResult struct {
    Orders []model.Order
    Cursor uint64       //this is where next time caller picks up from when FindAll() method is called
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
    res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

    keys, cursor, err := res.Result()
    if err != nil {
        return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
    }

    if len(keys) == 0 {
        return FindResult{
            Orders: []model.Order{},
        }, nil
    }

    //here we retrive individual values based on previously pulled out keys, Remember retrived data is unordered as we used a Set to store it
    xs, err := r.Client.MGet(ctx, keys...).Result()
    if err != nil {
        return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
    }

    orders := make([]model.Order, len(xs))

    //iterate thr xs, convert each element to string then decode the json data to model.Order type and then add each Unmarshalled order to orders slice
    for i, x := range xs {
        x := x.(string)
        var order model.Order

        err := json.Unmarshal([]byte(x), &order)
        if err != nil {
            return FindResult{}, fmt.Errorf("failed to decode order json: %w", err)
        }

        orders[i] = order
    }

    return FindResult{
        Orders: orders,
        Cursor: cursor,
    }, nil
}
