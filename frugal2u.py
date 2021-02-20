# -*-coding: utf-8 -*-

'''
实现 "Frugal Streaming for Estimating Quantiles" (Qiang Ma, S. Muthukrishnan, and Mark Sandler 2013)
( http://arxiv.org/abs/1407.1121 ) 中的算法 3 。

相关：http://blog.aggregateknowledge.com/2013/09/16/sketch-of-the-day-frugal-streaming/
'''

import random

class Frugal2U:

    def __init__(self, quantile):
        self.step = 1
        self.sign = 0
        self.m = None
        self.q = quantile

    def insert(self, s):
        if self.sign == 0:
            # 第一个元素作为初始估计
            self.m = s
            self.sign = 1
            return

        rand = random.random()

        if s > self.m and rand > 1 - self.q:
            self.step += self.sign * self.step
            self.m += self.step if self.step > 0 else 1
            if self.m > s:
                self.step += s - self.m
                self.m = s
            if self.sign < 0 and self.step > 1:
                self.step = 1
            self.sign = 1

        elif s < self.m and rand > self.q:
            self.step += -self.sign * self.step
            self.m -= self.step if self.step > 0 else 1
            if self.m < s:
                self.step += self.m - s
                self.m = s
            if self.sign > 0 and self.step > 1:
                self.step = 1
            self.sign = -1

    def estimate(self):
        return self.m

if __name__ == '__main__':
    S = [random.randint(1, 10000) for i  in range(1000)]
    print S
    f2 = Frugal2U(0.5)
    for s in S:
        f2.insert(s)
    print 'frugal2u:', f2.estimate()
    import numpy as np
    print 'exact:', np.percentile(S, 50)
