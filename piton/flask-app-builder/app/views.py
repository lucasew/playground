from flask import render_template, redirect, request, make_response, flash
from flask_appbuilder.models.sqla.interface import SQLAInterface
from flask_appbuilder import ModelView, ModelRestApi
from flask_appbuilder.forms import DynamicForm
from flask_appbuilder.actions import action
from flask_appbuilder.urltools import get_filter_args
from flask_appbuilder.baseviews import expose, expose_api
from flask_appbuilder.security.decorators import has_access_api, has_access, permission_name
from flask_appbuilder.widgets import FormVerticalWidget
from wtforms import Form
from wtforms.fields import FileField, StringField, Label

from . import appbuilder, db, models

class CounterApi(ModelRestApi):
    datamodel = SQLAInterface(models.Counter)

appbuilder.add_api(CounterApi)

class FileForm(DynamicForm):
    file = FileField(Label('file', "Arquivo"))
    name = StringField(Label('name', 'Nome do arquivo (se vazio vai ser detectado)'))

class VersionedFileStorage(ModelView):
    datamodel = SQLAInterface(models.VersionedFileStorage)

    label_columns = {
        'name': "Nome do arquivo",
        'version': 'Versão do arquivo',
        'blob': 'Arquivo',
        'blob_id': 'SHA256'
    }


    add_columns = ['file']
    edit_columns = ['name', 'file']
    list_columns = ['name', 'version', 'download']
    show_fieldsets = [
        ('Dados do arquivo', {'fields': ['name', 'version', 'download', 'blob_id']})
    ]

    @expose("/download/<pk>")
    @has_access
    def download(self, pk):
        item = self.datamodel.get(pk)
        response = make_response(item.blob.data, 200)
        response.headers['Content-Disposition'] = f"attachment; filename={item.name}"
        return response

    def _add(self): # https://github.com/dpgaspar/Flask-AppBuilder/blob/fab9013003a41c4e80da04f072201a8c7cc99187/flask_appbuilder/baseviews.py#L1208
        is_valid_form = True
        get_filter_args(self._filters, disallow_if_not_in_search=False)
        exclude_cols = self._filters.get_relation_cols()
        form = self.add_form.refresh()
        if request.method == 'POST':
            self._fill_form_exclude_cols(exclude_cols, form)
            if form.validate():
                self.process_form(form, True)
                item = self.datamodel.obj()

                try:
                    # print(form.file)
                    file = form.file.data

                    # print(dir(file))
                    # print(file.short_name)
                    # print('data', data)
                    # print('data', dir(data))
                    # print('data', data.filename)
                    # print('data', data.stream, type(data.stream))

                    item.name = "" if 'name' not in form else form.name.data
                    if item.name == "":
                        item.name = file.filename
                    item.blob = models.DataBlob(
                        data=file
                    )

                    self.pre_add(item)
                except Exception as e:
                    flash(str(e), "danger")
                    print(e)
                else:
                    if self.datamodel.add(item):
                        self.post_add(item)
                    flash(*self.datamodel.message)
                finally:
                    return None
            else:
                is_valid_form = False
        if is_valid_form:
            self.update_redirect()
        return self._get_add_widget(form=form, exclude_cols=exclude_cols)


    def _edit(self, pk): # https://github.com/dpgaspar/Flask-AppBuilder/blob/fab9013003a41c4e80da04f072201a8c7cc99187/flask_appbuilder/baseviews.py#L1208
        is_valid_form = True

        exclude_cols = self._filters.get_relation_cols()
        item = self.datamodel.get(pk, self._base_filters)        
        if not item:
            abort(404)
        pk = self.datamodel.get_pk_value(item)

        if request.method == 'POST':
            form = self.edit_form.refresh(request.form)
            self._fill_form_exclude_cols(exclude_cols, form)
            form._id = pk
            if form.validate():
                self.process_form(form, False)
                try:
                    file = form.file.data
                    # cria nova versão ao invés de alterar in-place
                    print('item_id a', item.id)
                    item = self.datamodel.obj()
                    print('item_id b', item.id)
                    item.blob = models.DataBlob(
                        data=file
                    )
                    old_name = item.name
                    item.name = "" if 'name' not in form else form.name.data
                    if item.name == "":
                        item.name = old_name
                    self.pre_update(item)
                except Exception as e:
                    flash(str(e), "danger")
                    print(e)
                else:
                    if self.datamodel.add(item):
                        self.post_update(item)
                    flash(*self.datamodel.message)
                finally:
                    return None
            else:
                is_valid_form = False
        else:
            form = self.edit_form.refresh()
            self.prefill_form(form, pk)
        print('formson', form.name.data)
        widgets = self._get_edit_widget(form, exclude_cols=exclude_cols)
        if is_valid_form:
          self.update_redirect()
        return widgets

    add_form = FileForm
    edit_form = FileForm

    def prefill_form(self, form, pk):
        print(pk)
        item = self.datamodel.get(pk)
        print(item.name)
        form.name.data = item.name

appbuilder.add_view(
    VersionedFileStorage,
    "Lista arquivos e versões",
    icon='fa-paperclip',
    category="Exemplos"
)

class FileStorageModelView(ModelView):
    datamodel = SQLAInterface(models.FileStorage)

    label_columns = {
        "name": "Nome do arquivo",
        "file": "O arquivo"
    }

    list_columns = ["name", "file"]


appbuilder.add_view(
    FileStorageModelView,
    "Lista arquivos",
    icon="fa-paperclip",
    category="Exemplos"
)


class CounterModelView(ModelView):
    datamodel = SQLAInterface(models.Counter)

    label_columns = {
        'name': "Nome do contador",
        'value': "Contagem"
    }
    list_columns = ["name", "value"]

    show_fieldsets = [
        ('Sumário', {'fields': ['name', 'value']})
    ]

    @action("increment", "Incrementar contador", "Dale?", "fa-plus")
    def increment(self, items):
        for item in items:
            item.value += 1
            db.session.add(item)
        db.session.commit()
        return redirect(self.get_redirect())

    @action("decrement", "Decrementar contador", "Dale?", "fa-minus")
    def decrement(self, items):
        for item in items:
            item.value -= 1
            db.session.add(item)
        db.session.commit()
        return redirect(self.get_redirect())

appbuilder.add_view(
    CounterModelView,
    "Lista contadores",
    icon="fa-list",
    category="Exemplos"
)


"""
    Create your Model based REST API::

    class MyModelApi(ModelRestApi):
        datamodel = SQLAInterface(MyModel)

    appbuilder.add_api(MyModelApi)


    Create your Views::


    class MyModelView(ModelView):
        datamodel = SQLAInterface(MyModel)


    Next, register your Views::


    appbuilder.add_view(
        MyModelView,
        "My View",
        icon="fa-folder-open-o",
        category="My Category",
        category_icon='fa-envelope'
    )
"""

"""
    Application wide 404 error handler
"""


@appbuilder.app.errorhandler(404)
def page_not_found(e):
    return (
        render_template(
            "404.html", base_template=appbuilder.base_template, appbuilder=appbuilder
        ),
        404,
    )


db.create_all()
