from flask_appbuilder import Model, expose, has_access, permission_name
from sqlalchemy import Column, Integer, String, ForeignKey, BLOB, event
from sqlalchemy.orm import relationship, Mapped
from flask_appbuilder.models.mixins import FileColumn
from flask_appbuilder.models.decorators import renders
from flask import Markup, url_for
from typing import List
import time
import hashlib

"""

You can use the extra Flask-AppBuilder fields and Mixin's

AuditMixin will add automatic timestamp of created and modified by who


"""

class Counter(Model):
    id = Column(Integer(), primary_key=True)
    name = Column(String(), unique=True, nullable=True)
    value = Column(Integer(), nullable=False, default=0)


class FileStorage(Model):
    id = Column(Integer(), primary_key=True)
    name = Column(String(), unique=True, nullable=True)
    file = Column(FileColumn, nullable=False)

    def download(self):
        return Markup(
            '<a target="_blank" href="'
            + url_for("ProjectFilesModelView.download", str(self.id))
            + '">Download</a>'
        )

    def file_name(self):
        return get_file_original_name(str(self.file))

class DataBlob(Model):
    hash = Column(String(), unique=True, nullable=False, primary_key=True)
    data = Column(BLOB, nullable=False)

    usages: Mapped[List['VersionedFileStorage']] = relationship('VersionedFileStorage', back_populates='blob')

@event.listens_for(DataBlob, 'before_insert')
def add_hash_to_blob(mapper, connect, target):
    if isinstance(target.data, str):
        target.data = target.data.encode('utf-8')
    if not isinstance(target.data, bytes):
        target.data = target.data.read()
    hasher = hashlib.sha256(target.data)
    target.hash = hasher.hexdigest()


class VersionedFileStorage(Model):
    id = Column(Integer(), primary_key=True)
    name = Column(String(), unique=True, nullable=False)
    version = Column(Integer(), nullable=False, default=lambda: int(time.time()))
    blob_id = Column(ForeignKey(DataBlob.hash))
    blob: Mapped['DataBlob'] = relationship('DataBlob', back_populates='usages')

    def download(self):
        return Markup(
            '<a target="_blank" href="'
            + url_for("VersionedFileStorage.download", pk=self.id)
            + '">Download</a>'
        )
