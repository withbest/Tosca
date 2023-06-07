#include "vm/evmzero/memory.h"

#include <gtest/gtest.h>

namespace tosca::evmzero {
namespace {

TEST(MemoryTest, DefaultInit) {
  Memory memory;
  EXPECT_EQ(memory.GetSize(), 0);
}

TEST(MemoryTest, InitializerList) {
  Memory memory = {1, 2, 3};
  EXPECT_EQ(memory.GetSize(), 3);

  EXPECT_EQ(memory[0], 1);
  EXPECT_EQ(memory[1], 2);
  EXPECT_EQ(memory[2], 3);
}

TEST(MemoryTest, ZeroInitialized) {
  Memory memory;
  memory.Grow(1);
  EXPECT_EQ(memory[0], 0);
}

TEST(MemoryTest, Grow) {
  Memory memory;
  EXPECT_EQ(memory.GetSize(), 0);

  memory.Grow(64);
  EXPECT_EQ(memory.GetSize(), 64);
}

TEST(MemoryTest, GrowRetainsElements) {
  Memory memory = {42};

  memory.Grow(2);
  EXPECT_EQ(memory[0], 42);
  EXPECT_EQ(memory[1], 0);
}

TEST(MemoryTest, GrowCanNotShrink) {
  Memory memory;
  memory.Grow(64);
  EXPECT_EQ(memory.GetSize(), 64);

  memory.Grow(32);
  EXPECT_EQ(memory.GetSize(), 64);
}

TEST(MemoryTest, SetMemory) {
  Memory memory;
  memory.SetMemory({1, 2, 3});
  EXPECT_EQ(memory.GetSize(), 3);

  EXPECT_EQ(memory[0], 1);
  EXPECT_EQ(memory[1], 2);
  EXPECT_EQ(memory[2], 3);
}

TEST(MemoryTest, ReadFrom) {
  Memory memory;

  std::vector<uint8_t> buffer = {1, 2, 3};
  memory.ReadFrom(buffer, 1);

  EXPECT_EQ(memory.GetSize(), 4);

  EXPECT_EQ(memory[0], 0);  // zero initialized
  EXPECT_EQ(memory[1], 1);
  EXPECT_EQ(memory[2], 2);
  EXPECT_EQ(memory[3], 3);
}

TEST(MemoryTest, ReadFromWithSize_SmallerSize) {
  Memory memory;

  std::vector<uint8_t> buffer = {1, 2, 3};
  memory.ReadFromWithSize(buffer, 1, 2);

  EXPECT_EQ(memory.GetSize(), 3);

  EXPECT_EQ(memory[0], 0);  // zero initialized
  EXPECT_EQ(memory[1], 1);
  EXPECT_EQ(memory[2], 2);
}

TEST(MemoryTest, ReadFromWithSize_LargerSize) {
  Memory memory = {0xFF, 0xFF, 0xFF, 0xFF, 0xFF};

  std::vector<uint8_t> buffer = {1, 2};
  memory.ReadFromWithSize(buffer, 1, 3);

  EXPECT_EQ(memory.GetSize(), 5);

  EXPECT_EQ(memory[0], 0xFF);
  EXPECT_EQ(memory[1], 1);
  EXPECT_EQ(memory[2], 2);
  EXPECT_EQ(memory[3], 0);  // filled with zero
  EXPECT_EQ(memory[4], 0xFF);
}

TEST(MemoryTest, WriteTo) {
  Memory memory;
  memory.SetMemory({1, 2, 3});

  std::vector<uint8_t> buffer(3);
  memory.WriteTo(buffer, 1);

  EXPECT_EQ(memory.GetSize(), 4);

  EXPECT_EQ(buffer[0], 2);
  EXPECT_EQ(buffer[1], 3);
  EXPECT_EQ(buffer[2], 0);  // zero initialized
}

TEST(MemoryTest, Subscript) {
  Memory memory = {1, 2, 3};

  memory[1] = 42;

  EXPECT_EQ(memory[0], 1);
  EXPECT_EQ(memory[1], 42);
  EXPECT_EQ(memory[2], 3);
}

TEST(MemoryTest, Equality) {
  Memory m1, m2;
  EXPECT_EQ(m1, m2);

  m1.Grow(3);
  EXPECT_NE(m1, m2);

  m2.SetMemory({1, 2, 3});
  EXPECT_NE(m1, m2);

  m1[0] = 1;
  m1[1] = 2;
  m1[2] = 3;
  EXPECT_EQ(m1, m2);
}

}  // namespace
}  // namespace tosca::evmzero