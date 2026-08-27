import 'dart:convert';
import 'package:http/http.dart' as http;

class EyeXPalette {
  final String background, surface, text, primary, secondary, error, success;
  const EyeXPalette({required this.background,required this.surface,required this.text,required this.primary,required this.secondary,required this.error,required this.success});
  factory EyeXPalette.fromJson(Map<String,dynamic> j)=>EyeXPalette(background:j['background'],surface:j['surface'],text:j['text'],primary:j['primary'],secondary:j['secondary'],error:j['error'],success:j['success']);
  Map<String,dynamic> toJson()=>{'background':background,'surface':surface,'text':text,'primary':primary,'secondary':secondary,'error':error,'success':success};
}
class EyeXTheme {
  final String type; final EyeXPalette palette; final bool contrastOk;
  const EyeXTheme({required this.type,required this.palette,required this.contrastOk});
  factory EyeXTheme.fromJson(Map<String,dynamic> j)=>EyeXTheme(type:j['type'],palette:EyeXPalette.fromJson(j['palette']),contrastOk:j['contrast_ok'] as bool);
}
class EyeXTestAnswers {
  final bool redsLookDarker, greenBrownConfusion, blueYellowConfusion, colorsLookGray;
  const EyeXTestAnswers({required this.redsLookDarker,required this.greenBrownConfusion,required this.blueYellowConfusion,required this.colorsLookGray});
  Map<String,dynamic> toJson()=>{'reds_look_darker':redsLookDarker,'green_brown_confusion':greenBrownConfusion,'blue_yellow_confusion':blueYellowConfusion,'colors_look_gray':colorsLookGray};
}
class EyeXTestResult {
  final String suggestedType, disclaimer;
  const EyeXTestResult(this.suggestedType,this.disclaimer);
  factory EyeXTestResult.fromJson(Map<String,dynamic> j)=>EyeXTestResult(j['suggested_type'],j['disclaimer']);
}
class EyeXClient {
  final String baseUrl;
  const EyeXClient(this.baseUrl);
  String get _base=>baseUrl.replaceFirst(RegExp(r'/$'),'');
  Future<Map<String,dynamic>> _json(Uri uri,{String method='GET',Map<String,dynamic>? body}) async {
    final headers={'Accept':'application/json',if(body!=null)'Content-Type':'application/json'};
    final response=method=='POST'?await http.post(uri,headers:headers,body:jsonEncode(body)):await http.get(uri,headers:headers);
    final data=jsonDecode(response.body) as Map<String,dynamic>;
    if(response.statusCode<200||response.statusCode>=300)throw Exception(data['message']??data['error']??'EyeX request failed');
    return data;
  }
  Future<List<String>> types() async {final data=await _json(Uri.parse('$_base/api/v1/theme/types'));return (data['types'] as List).cast<String>();}
  Future<EyeXTheme> theme(String type,{String? severity,String? mode,bool? highContrast}) async {
    final params=<String,String>{};if(severity!=null)params['severity']=severity;if(mode!=null)params['mode']=mode;if(highContrast!=null)params['high_contrast']='$highContrast';
    final uri=Uri.parse('$_base/api/v1/theme/$type').replace(queryParameters:params.isEmpty?null:params);
    return EyeXTheme.fromJson(await _json(uri));
  }
  Future<EyeXTheme> custom(String type,EyeXPalette palette,{String? severity,String? mode,bool? highContrast}) async {
    final body=<String,dynamic>{'type':type,'palette':palette.toJson(),if(severity!=null)'severity':severity,if(mode!=null)'mode':mode,if(highContrast!=null)'high_contrast':highContrast};
    return EyeXTheme.fromJson(await _json(Uri.parse('$_base/api/v1/theme/custom'),method:'POST',body:body));
  }
  Future<EyeXTestResult> suggest(EyeXTestAnswers answers) async {
    final data=await _json(Uri.parse('$_base/api/v1/test/suggest'),method:'POST',body:{'answers':answers.toJson()});
    return EyeXTestResult.fromJson(data);
  }
}
