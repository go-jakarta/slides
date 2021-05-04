precision highp float;

#include <lights>
#include <material>
#include <phong_model>

// Inputs from vertex shader
in vec4 Position;       // Vertex position in camera coordinates.
in vec3 Normal;         // Vertex normal in camera coordinates.
in vec3 CamDir;         // Direction from vertex to camera
in vec2 FragTexcoord;

in vec4 worldPosition;

// Final fragment color
out vec4 FragColor;

void logisticInterp(vec4 a, vec4 b, float f, out float r) {

}

void main() {

    vec4 texDay = texture(MatTexture[0], FragTexcoord * MatTexRepeat(0) + MatTexOffset(0));
    vec4 texSpecular = texture(MatTexture[1], FragTexcoord * MatTexRepeat(1) + MatTexOffset(1));
    vec4 texNight = texture(MatTexture[2], FragTexcoord * MatTexRepeat(2) + MatTexOffset(2));

    vec3 sunDirection = normalize(DirLightPosition(0));

    //vec4 texDayOrNight;// = texDay;

    // Inverts the fragment normal if not FrontFacing
    vec3 fragNormal = Normal;
    if (!gl_FrontFacing) {
        fragNormal = -fragNormal;
    }

    float dotNormal = dot(sunDirection, fragNormal);
    //if (dotNormal < 0) {
    //	texDayOrNight = texNight;
    //}

    vec4 texDayOrNight = mix(texNight, texDay, max(min((((dotNormal + 1.0)/2.0) - 0.45)*10.0, 1.0), 0.0)  );

    // Combine material with texture colors
    vec4 matDiffuse = vec4(MatDiffuseColor, MatOpacity) * texDayOrNight;
    vec4 matAmbient = vec4(MatAmbientColor, MatOpacity) * texDayOrNight;

    // Calculates the Ambient+Diffuse and Specular colors for this fragment using the Phong model.
    vec3 Ambdiff, Spec;
    phongModel(Position, fragNormal, CamDir, vec3(matAmbient), vec3(matDiffuse), Ambdiff, Spec);

    // Calculate specular mask
    Spec = vec3(texSpecular) * Spec;

    // Final fragment color
    FragColor = min(vec4(Ambdiff + Spec, matDiffuse.a), vec4(1.0));
}
