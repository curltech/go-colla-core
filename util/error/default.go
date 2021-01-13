package error

import "log"

func Try(f func() error) (err error) {
	// 错误拦截必须配合defer使用，通过匿名函数使用，在错误之前引用
	defer func() {
		err := recover()
		if err != nil {
			log.Panicf("exception catched")
		}
	}()
	// 执行在事务内的处理
	err = f()

	return err
}
