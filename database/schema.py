from operator import index
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy import (
    MetaData,
    Column,
    Text,
    Integer,
    DateTime,
    UniqueConstraint
)

NAMING_CONVENTION = {
  "ix": 'ix_%(column_0_label)s',
  "uq": "uq_%(table_name)s_%(column_0_N_name)s",
  "ck": "ck_%(table_name)s_%(constraint_name)s",
  "fk": "fk_%(table_name)s_%(column_0_name)s_%(referred_table_name)s",
  "pk": "pk_%(table_name)s"
}

metadata = MetaData(naming_convention=NAMING_CONVENTION)
Base = declarative_base(metadata=metadata)

class User(Base):
  __tablename__ = 'user'

  uuid = Column(UUID, primary_key=True)
  email = Column(Text, nullable=False, unique=True)
  password = Column(Text, nullable=False)
  name = Column(Text, nullable=False)
  session = Column(Text, nullable=True)

class Product(Base):
  __tablename__ = 'product'
  
  uuid = Column(UUID, primary_key=True)
  name = Column(Text, nullable=False)
  description = Column(Text, nullable=False)
  stock = Column(Integer, nullable=False)
  price = Column(Integer, nullable=False)

class Transaction(Base):
  __tablename__ = 'transaction'

  __table_args__ = (
    UniqueConstraint('payment_method', 'virtual_account'),
  )

  uuid = Column(UUID, primary_key=True)
  user_uuid = Column(UUID, nullable=False, index=True)
  status = Column(Text, nullable=False)
  amount = Column(Integer, nullable=False)
  created = Column(DateTime, nullable=False)
  expired = Column(DateTime, nullable=False)
  payment_method = Column(Text, nullable=False)
  virtual_account = Column(Text, nullable=False)

class TransactionCart(Base):
  __tablename__ = 'transaction_cart'

  uuid = Column(UUID, primary_key=True)
  product_uuid = Column(UUID, nullable=False, index=True)
  user_uuid = Column(UUID, nullable=False, index=True)
  qty = Column(Integer, nullable=False)
  transaction_uuid = Column(UUID, nullable=True)

class TransactionPayment(Base):
  __tablename__ = 'transaction_payment'

  uuid = Column(UUID, primary_key=True)
  reference = Column(Text, nullable=False)
  amount = Column(Integer, nullable=False)
  transaction_uuid = Column(UUID, nullable=False)
  payment_datetime = Column(DateTime, nullable=False)


class Order(Base):
  __tablename__ = 'order'

  uuid = Column(UUID, primary_key=True)
  user_uuid = Column(UUID, nullable=False, index=True)
  amount = Column(Integer, nullable=False)
  created = Column(DateTime, nullable=False)

class OrderItem(Base):
  __tablename__ = 'order_item'

  order_uuid = Column(UUID, primary_key=True)
  product_uuid = Column(UUID, nullable=False, primary_key=True)
  name = Column(Text, nullable=False)
  description = Column(Text, nullable=False)
  price = Column(Integer, nullable=False)
  qty = Column(Integer, nullable=False)