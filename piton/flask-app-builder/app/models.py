from flask_appbuilder import Model
from sqlalchemy import Column, Integer, String, ForeignKey
from sqlalchemy.orm import relationship

"""

You can use the extra Flask-AppBuilder fields and Mixin's

AuditMixin will add automatic timestamp of created and modified by who


"""

class Counter(Model):
    id = Column(Integer(), primary_key=True)
    name = Column(String(), unique=True, nullable=True)
    value = Column(Integer(), nullable=False, default=0)
