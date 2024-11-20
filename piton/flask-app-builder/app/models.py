from flask_appbuilder import Model
from sqlalchemy import Column, Integer, String, ForeignKey
from sqlalchemy.orm import relationship
from flask_appbuilder.models.mixins import FileColumn

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
            '<a href="'
            + url_for("ProjectFilesModelView.download", filename=str(self.file))
            + '">Download</a>'
        )

    def file_name(self):
        return get_file_original_name(str(self.file))
