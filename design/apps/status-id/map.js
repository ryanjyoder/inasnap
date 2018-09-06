function (doc) {
        var status = "";
    if( doc.status){
        status = doc.status;
    }
    emit(status, doc._id);
}